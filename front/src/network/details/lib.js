import {request} from "../request";

export function create(data) {
    return request({
        url: '/api/libs',
        method: 'post',
        data: data,
    })
}

export function update(libId, data) {
    return request({
        url: `/api/libs/${libId}`,
        method: 'put',
        data: data,
    })
}

export function getListSelf() {
    return request({
        url: '/api/libs?scope=self',
        method: 'get',
    })
}

export function getLibInfo(libId) {
    return request({
        url: `/api/libs/${libId}`,
        method: 'get',
    })
}

export function getListCommon({pageNum, pageSize, q}) {
    return request({
        url: '/api/libs?scope=common',
        method: 'get',
        params: {pageNum, pageSize, q},
        animation: false
    })
}

export function add(libId) {
    return request({
        url: `/api/libs/${libId}/adopt`,
        method: 'post',
        animation: false
    })
}

export function remove(libId) {
    return request({
        url: `/api/libs/${libId}`,
        method: 'delete',
        animation: false
    })
}
