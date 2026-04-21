import { request } from './http'
import type { AppealListItem, AuditAppealPayload, AuditReviewPayload, ListReply, PaginationParams, ReviewListItem } from '../types/review'

export function auditReview(payload: AuditReviewPayload) {
  return request<void>('operation', {
    method: 'POST',
    url: '/v1/op/review/audit',
    data: payload,
  })
}

export function auditAppeal(payload: AuditAppealPayload) {
  return request<void>('operation', {
    method: 'POST',
    url: '/v1/op/review/appeal/audit',
    data: payload,
  })
}

export function listPendingReviews(params: PaginationParams) {
  return request<ListReply<ReviewListItem>>('operation', {
    method: 'GET',
    url: '/v1/op/review/pending',
    params,
  })
}

export function listPendingAppeals(params: PaginationParams) {
  return request<ListReply<AppealListItem>>('operation', {
    method: 'GET',
    url: '/v1/op/review/appeal/pending',
    params,
  })
}
