import { request } from './http'
import type { ReplyReviewPayload } from '../types/review'

export function replyReview(payload: ReplyReviewPayload) {
  return request<{ replyID: number }>('business', {
    method: 'POST',
    url: '/business/v1/review/reply',
    data: payload,
  })
}
