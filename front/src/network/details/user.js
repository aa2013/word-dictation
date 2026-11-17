import {request} from "../request";

export function me() {
    return request({
        url: '/api/auth/me',
        method: 'get',
        animation: false,
        dialog: false
    })
}

export function login(data) {
    return request({
        url: '/api/auth/login',
        method: 'post',
        data: data,
        animation: false
    })
}

export function logout() {
    return request({
        url: '/api/auth/logout',
        method: 'post',
        animation: false,
        dialog: false
    })
}

export function register(data) {
    return request({
        url: '/api/users',
        method: 'post',
        data: data,
        animation: false
    })
}

export function sendRegisterCode(data) {
    return request({
        url: '/api/users/codes/register',
        method: 'post',
        data: data,
        animation: false
    })
}

export function sendForgetPwdCode(data) {
    return request({
        url: '/api/users/codes/reset',
        method: 'post',
        data: data,
        animation: false
    })
}

export function resetPwd(data) {
    return request({
        url: '/api/users/password/reset',
        method: 'post',
        data: data,
        animation: false
    })
}

export function updatePwd(data) {
    return request({
        url: '/api/users/password',
        method: 'put',
        data: data
    })
}
