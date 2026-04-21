export type Int64Value = string | number

export interface ReviewListItem {
  reviewID: number
  userID: number
  orderID: number
  score: number
  serviceScore: number
  expressScore: number
  content: string
  picInfo: string
  videoInfo: string
  status: number
  hasReply: number
  createAt: number
}

export interface AppealListItem {
  appealID: Int64Value
  reviewID: Int64Value
  storeID: Int64Value
  status: number
  reason: string
  content: string
  picInfo: string
  videoInfo: string
  opRemarks: string
  opUser: string
  createAt: Int64Value
}

export interface GoodsScoreRankItem {
  spuID: Int64Value
  avgScore: number
  reviewCount: Int64Value
}

export interface UploadedMediaItem {
  url: string
  name?: string
  size?: number
}

export interface UploadImageReply {
  url: string
  name: string
  size: number
}

export interface ListReply<T> {
  list: T[]
  total: number
}

export interface CreateReviewPayload {
  userID: number
  orderID: number
  storeID: number
  score: number
  serviceScore: number
  expressScore: number
  content: string
  picInfo: string
  videoInfo: string
  anonymous: boolean
}

export interface UpdateReviewPayload {
  reviewID: number
  userID: number
  score: number
  serviceScore: number
  expressScore: number
  content: string
  picInfo: string
  videoInfo: string
}

export interface DeleteReviewPayload {
  reviewID: number
  userID: number
}

export interface ReplyReviewPayload {
  reviewID: Int64Value
  storeID: number
  content: string
  picInfo: string
  videoInfo: string
  extJSON: string
  ctrlJSON: string
}

export interface AppealReviewPayload {
  userID: number
  reviewID: Int64Value
  reason: string
  picInfo: string
}

export interface AuditReviewPayload {
  reviewID: Int64Value
  result: number
  remark: string
  operator: string
}

export interface AuditAppealPayload {
  appealID: Int64Value
  result: number
  remark: string
  operator: string
}

export interface PaginationParams {
  page: number
  pageSize: number
}
