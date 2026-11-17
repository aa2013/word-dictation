import {request} from "../request";

export function add(data) {
    return request({
        url: '/api/plans',
        method: 'post',
        data: data,
    })
}

export function getList(data) {
    return request({
        url: '/api/plans',
        method: 'get',
        params: data,
    })
}

export function getWords(planId) {
    return request({
        url: `/api/plans/${planId}/words`,
        method: 'get',
    })
}

export function remove(planId) {
    return request({
        url: `/api/plans/${planId}`,
        method: 'delete',
    })
}
