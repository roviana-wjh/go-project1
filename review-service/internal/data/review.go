package data

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"review-service/internal/biz"
	"review-service/internal/data/model"
	"review-service/internal/data/query"
	"review-service/pkg/snowflake"

	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/sortorder"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/singleflight"
)

// #region agent log
// agentDebugLogPathResolved writes next to repo workspace root (parent of review-service/) when possible,
// so logs are visible under d:\go\rpc\new\debug-b3073d.log regardless of process cwd (e.g. cmd/).
func agentDebugLogPathResolved() string {
	if p := os.Getenv("DEBUG_AGENT_LOG"); p != "" {
		return p
	}
	wd, err := os.Getwd()
	if err != nil {
		return "debug-b3073d.log"
	}
	dir := wd
	for range 16 {
		if filepath.Base(dir) == "review-service" {
			if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
				return filepath.Join(filepath.Dir(dir), "debug-b3073d.log")
			}
		}
		if _, err := os.Stat(filepath.Join(dir, "review-service", "go.mod")); err == nil {
			return filepath.Join(dir, "debug-b3073d.log")
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return filepath.Join(wd, "debug-b3073d.log")
}

func agentDebugNDJSON(hypothesisID, location, message string, data map[string]any) {
	b, _ := json.Marshal(map[string]any{
		"sessionId":    "b3073d",
		"hypothesisId": hypothesisID,
		"location":     location,
		"message":      message,
		"data":         data,
		"timestamp":    time.Now().UnixMilli(),
		"runId":        "post-fix",
	})
	line := append(b, '\n')
	paths := []string{
		agentDebugLogPathResolved(),
		"debug-b3073d.log",
		filepath.Join(os.TempDir(), "debug-b3073d.log"),
	}
	for _, p := range paths {
		if f, err := os.OpenFile(p, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
			_, werr := f.Write(line)
			_ = f.Close()
			if werr == nil {
				return
			}
		}
	}
	_, _ = fmt.Fprintln(os.Stderr, string(b))
}

// #endregion

// storeReviewListCache Redis 缓存一页列表（与 storeReviewsFromES 语义一致）
type storeReviewListCache struct {
	Total   int64                 `json:"total"`
	Reviews []*model.ReviewInfo   `json:"reviews"`
}

type ReviewRepo struct {
	data *Data
	log  *log.Helper
}

func NewReviewRepo(data *Data, logger log.Logger) biz.ReviewRepo {
	return &ReviewRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r *ReviewRepo) SaveReview(ctx context.Context, review *model.ReviewInfo) (*model.ReviewInfo, error) {
	if err := r.data.q.ReviewInfo.WithContext(ctx).Create(review); err != nil {
		return nil, wrapReviewDB("创建评价", err)
	}
	return review, nil
}

func (r *ReviewRepo) GetReviewByOrderID(ctx context.Context, orderID int64) ([]*model.ReviewInfo, error) {
	reviews, err := r.data.q.ReviewInfo.WithContext(ctx).Where(r.data.q.ReviewInfo.OrderID.Eq(orderID)).Find()
	if err != nil {
		return nil, wrapReviewDB("按订单查询评价", err)
	}
	return reviews, nil
}

func (r *ReviewRepo) GetByReviewID(ctx context.Context, reviewID int64) (*model.ReviewInfo, error) {
	row, err := r.data.q.ReviewInfo.WithContext(ctx).Where(r.data.q.ReviewInfo.ReviewID.Eq(reviewID)).First()
	if err != nil {
		return nil, wrapReviewDB("查询评价", err)
	}
	return row, nil
}

func (r *ReviewRepo) ListByOrderID(ctx context.Context, p *biz.ReviewListOrderParams) ([]*model.ReviewInfo, int64, error) {
	q := r.data.q.ReviewInfo.WithContext(ctx).Where(r.data.q.ReviewInfo.OrderID.Eq(p.OrderID))
	offset := int((p.Page - 1) * p.PageSize)
	list, total, err := q.Order(r.data.q.ReviewInfo.CreateAt.Desc()).FindByPage(offset, int(p.PageSize))
	if err != nil {
		return nil, 0, wrapReviewDB("分页查询评价", err)
	}
	return list, total, nil
}

func (r *ReviewRepo) UpdateReview(ctx context.Context, row *model.ReviewInfo) error {
	if err := r.data.q.ReviewInfo.WithContext(ctx).Save(row); err != nil {
		return wrapReviewDB("更新评价", err)
	}
	return nil
}

func (r *ReviewRepo) DeleteByReviewID(ctx context.Context, reviewID int64) error {
	info, err := r.data.q.ReviewInfo.WithContext(ctx).Where(r.data.q.ReviewInfo.ReviewID.Eq(reviewID)).First()
	if err != nil {
		return wrapReviewDB("删除前查询评价", err)
	}
	if _, err := r.data.q.ReviewInfo.WithContext(ctx).Delete(info); err != nil {
		return wrapReviewDB("删除评价", err)
	}
	return nil
}
func (r *ReviewRepo) SaveReply(ctx context.Context, reply *model.ReviewReplyInfo) (*model.ReviewReplyInfo, error) {
	//数据校验
	//1. 已回复的评价不允许商家再次回复
	review, err := r.data.q.ReviewInfo.WithContext(ctx).Where(r.data.q.ReviewInfo.ReviewID.Eq(reply.ReviewID)).First()
	if err != nil {
		return nil, errors.New("评价不存在")
	}
	if review.HasReply == 1 {
		return nil, errors.New("已回复的评价不允许商家再次回复")
	}

	//2. 水平越权校验 A-B 权限
	if review.StoreID != reply.StoreID {
		return nil, errors.New("无权回复他人评价")
	}
	// 3. 同一事务：写入回复 + 将评价 has_reply 置为 1
	review.HasReply = 1
	if err := r.data.q.Transaction(func(tx *query.Query) error {
		if err := tx.ReviewReplyInfo.WithContext(ctx).Save(reply); err != nil {
			return err
		}
		if err := tx.ReviewInfo.WithContext(ctx).Where(tx.ReviewInfo.ReviewID.Eq(review.ReviewID)).Save(review); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, wrapReviewDB("保存商家回复", err)
	}
	return reply, nil
}

func (r *ReviewRepo) GetByAppealID(ctx context.Context, appealID int64) (*model.ReviewAppealInfo, error) {
	row, err := r.data.q.ReviewAppealInfo.WithContext(ctx).Where(r.data.q.ReviewAppealInfo.AppealID.Eq(appealID)).First()
	if err != nil {
		return nil, wrapReviewDB("查询申诉", err)
	}
	return row, nil
}

func (r *ReviewRepo) ListByUseId(ctx context.Context, p *biz.ReviewListUserParams) ([]*model.ReviewInfo, int64, error) {
	q := r.data.q.ReviewInfo.WithContext(ctx).Where(r.data.q.ReviewInfo.UserID.Eq(p.UserID))
	offset := int((p.Page - 1) * p.PageSize)
	list, total, err := q.Order(r.data.q.ReviewInfo.CreateAt.Desc()).FindByPage(offset, int(p.PageSize))
	if err != nil {
		return nil, 0, wrapReviewDB("分页查询评价", err)
	}
	return list, total, nil
}
func (r *ReviewRepo) SaveAppeal(ctx context.Context, params *biz.AppealReviewParams) (*model.ReviewAppealInfo, error) {
	//先查询有没有申诉记录
	row, err := r.data.q.ReviewAppealInfo.WithContext(ctx).Where(r.data.q.ReviewAppealInfo.ReviewID.Eq(params.ReviewID),
		r.data.q.ReviewAppealInfo.StoreID.Eq(params.StoreID)).First()
	if err != nil {
		return nil, wrapReviewDB("查询申诉", err)
	}
	if err == nil && row.Status > 10 {
		return nil, errors.New("已存在审核过的申诉记录，不允许再次申诉")
	}
	//有申诉记录且处于待审核状态，则更新申诉
	appeal := &model.ReviewAppealInfo{
		AppealID: snowflake.GenID(),
		ReviewID: params.ReviewID,
		StoreID:  params.StoreID,
		Reason:   params.Reason,
		PicInfo:  params.PicInfo,
		CreateBy: fmt.Sprintf("%d", params.UserID),
		UpdateBy: fmt.Sprintf("%d", params.UserID),
	}
	if row != nil {
		r.data.q.ReviewAppealInfo.WithContext(ctx).Where(r.data.q.ReviewAppealInfo.AppealID.Eq(row.AppealID)).Save(appeal)
	} else {
		r.data.q.ReviewAppealInfo.WithContext(ctx).Create(appeal)
	}
	return appeal, nil
}
func (r *ReviewRepo) AuditAppeal(ctx context.Context, p *biz.AuditAppealParams) (*model.ReviewAppealInfo, error) {
	row, err := r.data.q.ReviewAppealInfo.WithContext(ctx).Where(r.data.q.ReviewAppealInfo.AppealID.Eq(p.AppealID),
		r.data.q.ReviewReplyInfo.StoreID.Eq(p.StoreID)).First()
	if err != nil {
		return nil, wrapReviewDB("查询申诉", err)
	}
	r.log.WithContext(ctx).Debugf("[data] audit appeal: appealID=%d storeID=%d", p.AppealID, p.StoreID)
	if row.Status > 10 {
		return nil, errors.New("仅待审核状态可审核")
	}
	appeal := &model.ReviewAppealInfo{
		AppealID: p.AppealID,
		ReviewID: row.ReviewID,
		StoreID:  row.StoreID,
		Reason:   row.Reason,
		PicInfo:  row.PicInfo,
		CreateBy: row.CreateBy,
		UpdateBy: p.Operator,
	}
	return appeal, nil
}
func (r *ReviewRepo) ListByStoreId(ctx context.Context, p *biz.ReviewListStoreParams) ([]*model.ReviewInfo, int64, error) {
	return r.getData2(ctx, p)
}

func (r *ReviewRepo) getData2(ctx context.Context, p *biz.ReviewListStoreParams) ([]*model.ReviewInfo, int64, error) {
	// #region agent log
	agentDebugNDJSON("H1", "review.go:getData2", "entry", map[string]any{
		"rdbNil":  r.data.rdb == nil,
		"storeId": p.StoreID,
		"page":    p.Page,
		"size":    p.PageSize,
	})
	// #endregion
	if r.data.rdb == nil {
		return r.storeReviewsFromES(ctx, p)
	}
	from := int((p.Page - 1) * p.PageSize)
	size := int(p.PageSize)
	// v2: JSON storeReviewListCache; old keys may hold incompatible payloads from previous implementations
	key := fmt.Sprintf("review:v2:store_id:%d:offset:%d:limit:%d", p.StoreID, from, size)
	raw, err := r.getDataBySingleflight(ctx, key, p)
	if err != nil {
		return nil, 0, err
	}
	var c storeReviewListCache
	if err := json.Unmarshal(raw, &c); err != nil {
		// #region agent log
		agentDebugNDJSON("H3", "review.go:getData2", "cache_unmarshal_fail", map[string]any{
			"key": key, "rawLen": len(raw), "err": err.Error(),
		})
		// #endregion
		return nil, 0, fmt.Errorf("店铺评价缓存解码: %w", err)
	}
	// #region agent log
	agentDebugNDJSON("H1", "review.go:getData2", "exit_ok", map[string]any{"total": c.Total, "len": len(c.Reviews)})
	// #endregion
	return c.Reviews, c.Total, nil
}

// storeReviewsFromES 直查 ES（与缓存中保存的结构一致）
func (r *ReviewRepo) storeReviewsFromES(ctx context.Context, p *biz.ReviewListStoreParams) ([]*model.ReviewInfo, int64, error) {
	if r.data.es == nil {
		r.log.WithContext(ctx).Errorf("[data] ListByStoreId: elasticsearch 未初始化")
		return nil, 0, fmt.Errorf("elasticsearch client 未初始化")
	}
	from := int((p.Page - 1) * p.PageSize)
	size := int(p.PageSize)
	desc := sortorder.Desc
	// Canal 同步多为字符串，动态映射常见为 text + .keyword；对 text 做 sort 会导致分片失败。
	sortOpts := types.SortOptions{
		SortOptions: map[string]types.FieldSort{
			"create_at.keyword": {Order: &desc},
		},
	}
	resp, err := r.data.es.Search().
		Index("review").
		From(from).
		Size(size).
		TrackTotalHits(true).
		Sort(sortOpts).
		Query(&types.Query{
			Bool: &types.BoolQuery{
				Filter: []types.Query{
					{
						Bool: &types.BoolQuery{
							Should: []types.Query{
								{Term: map[string]types.TermQuery{
									"store_id": {Value: p.StoreID},
								}},
								{Term: map[string]types.TermQuery{
									"store_id.keyword": {Value: strconv.FormatInt(p.StoreID, 10)},
								}},
							},
							MinimumShouldMatch: 1,
						},
					},
				},
			},
		}).
		Do(ctx)
	if err != nil {
		r.log.WithContext(ctx).Errorf("[data] ListByStoreId ES 查询失败 store_id=%d from=%d size=%d: %v", p.StoreID, from, size, err)
		return nil, 0, fmt.Errorf("按店铺查询评价(ES): %w", err)
	}
	var total int64
	if resp.Hits.Total != nil {
		total = resp.Hits.Total.Value
	}
	out := make([]*model.ReviewInfo, 0, len(resp.Hits.Hits))
	for i := range resp.Hits.Hits {
		row, err := reviewInfoFromESSource(resp.Hits.Hits[i].Source_)
		if err != nil {
			r.log.WithContext(ctx).Warnf("跳过无法解析的 ES 文档: %v", err)
			continue
		}
		out = append(out, row)
	}
	r.log.WithContext(ctx).Debugf("[data] ListByStoreId store_id=%d es_total=%d parsed=%d", p.StoreID, total, len(out))
	return out, total, nil
}

func reviewInfoFromESSource(src json.RawMessage) (*model.ReviewInfo, error) {
	if len(src) == 0 {
		return nil, fmt.Errorf("empty _source")
	}
	dec := json.NewDecoder(bytes.NewReader(src))
	dec.UseNumber()
	var m map[string]any
	if err := dec.Decode(&m); err != nil {
		return nil, err
	}
	return &model.ReviewInfo{
		ID:             esInt64(m, "id"),
		CreateBy:       esString(m, "create_by"),
		UpdateBy:       esString(m, "update_by"),
		CreateAt:       esTime(m, "create_at"),
		UpdateAt:       esTime(m, "update_at"),
		DeleteAt:       esTimePtr(m, "delete_at"),
		Version:        int32(esInt64(m, "version")),
		ReviewID:       esInt64(m, "review_id"),
		Content:        esString(m, "content"),
		Score:          int32(esInt64(m, "score")),
		ServiceScore:   int32(esInt64(m, "service_score")),
		ExpressScore:   int32(esInt64(m, "express_score")),
		HasMedia:       int32(esInt64(m, "has_media")),
		OrderID:        esInt64(m, "order_id"),
		SkuID:          esInt64(m, "sku_id"),
		SpuID:          esInt64(m, "spu_id"),
		StoreID:        esInt64(m, "store_id"),
		UserID:         esInt64(m, "user_id"),
		Anonymous:      int32(esInt64(m, "anonymous")),
		Tags:           esString(m, "tags"),
		PicInfo:        esString(m, "pic_info"),
		VideoInfo:      esString(m, "video_info"),
		Status:         int32(esInt64(m, "status")),
		IsDefault:      int32(esInt64(m, "is_default")),
		HasReply:       int32(esInt64(m, "has_reply")),
		OpReason:       esString(m, "op_reason"),
		OpRemarks:      esString(m, "op_remarks"),
		OpUser:         esString(m, "op_user"),
		GoodsSnapshoot: esString(m, "goods_snapshoot"),
		ExtJSON:        esString(m, "ext_json"),
		CtrlJSON:       esString(m, "ctrl_json"),
	}, nil
}

func esString(m map[string]any, k string) string {
	v, ok := m[k]
	if !ok || v == nil {
		return ""
	}
	switch x := v.(type) {
	case string:
		return x
	case json.Number:
		return x.String()
	case float64:
		return strconv.FormatInt(int64(x), 10)
	default:
		return fmt.Sprint(v)
	}
}

func esInt64(m map[string]any, k string) int64 {
	v, ok := m[k]
	if !ok || v == nil {
		return 0
	}
	switch x := v.(type) {
	case json.Number:
		i, _ := x.Int64()
		return i
	case float64:
		return int64(x)
	case string:
		i, _ := strconv.ParseInt(x, 10, 64)
		return i
	case int64:
		return x
	default:
		return 0
	}
}

func esTime(m map[string]any, k string) time.Time {
	s := esString(m, k)
	if s == "" {
		return time.Time{}
	}
	for _, layout := range []string{time.RFC3339Nano, time.RFC3339, "2006-01-02 15:04:05"} {
		if t, err := time.ParseInLocation(layout, s, time.Local); err == nil {
			return t
		}
	}
	return time.Time{}
}

func esTimePtr(m map[string]any, k string) *time.Time {
	if _, ok := m[k]; !ok {
		return nil
	}
	t := esTime(m, k)
	if t.IsZero() {
		return nil
	}
	return &t
}

var g singleflight.Group

// 带缓存版本的查询：缓存值为 JSON(storeReviewListCache)；未命中时与 storeReviewsFromES 一致
func (r *ReviewRepo) getDataBySingleflight(ctx context.Context, key string, p *biz.ReviewListStoreParams) ([]byte, error) {
	v, err, shared := g.Do(key, func() (any, error) {
		data, err := r.getDataFromCahe(ctx, key)
		r.log.WithContext(ctx).Debugf("[data] getDataFromCahe: data=%v err=%v", data, err)
		if err == nil {
			// #region agent log
			agentDebugNDJSON("H2", "review.go:getDataBySingleflight", "cache_hit", map[string]any{"key": key})
			// #endregion
			return data, nil
		}
		if errors.Is(err, redis.Nil) {
			reviews, total, err := r.storeReviewsFromES(ctx, p)
			if err != nil {
				return nil, err
			}
			c := storeReviewListCache{Total: total, Reviews: reviews}
			raw, err := json.Marshal(&c)
			if err != nil {
				return nil, err
			}
			if err := r.setCache(ctx, key, raw, 10*time.Minute); err != nil {
				r.log.WithContext(ctx).Warnf("[data] setCache: %v", err)
			}
			// #region agent log
			agentDebugNDJSON("H2", "review.go:getDataBySingleflight", "cache_fill", map[string]any{
				"key": key, "total": total, "n": len(reviews),
			})
			// #endregion
			return raw, nil
		}
		return nil, err
	})
	r.log.WithContext(ctx).Debugf("[data] getDataBySingleflight: key=%s v=%v err=%v shared=%v", key, v, err, shared)
	if err != nil {
		return nil, err
	}
	return v.([]byte), nil
}

// 读缓存
func (r *ReviewRepo) getDataFromCahe(ctx context.Context, key string) (any, error) {
	r.log.WithContext(ctx).Debugf("[data] getDataFromCahe: key=%s", key)
	return r.data.rdb.Get(ctx, key).Bytes()
}

// 写缓存
func (r *ReviewRepo) setCache(ctx context.Context, key string, data any, ttl time.Duration) error {
	r.log.WithContext(ctx).Debugf("[data] setCache: key=%s data=%v ttl=%s", key, data, ttl)
	return r.data.rdb.Set(ctx, key, data, ttl).Err()
}
