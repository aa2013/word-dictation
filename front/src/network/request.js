import axios from "axios";
import Vue from 'vue'

// 统一请求封装：响应包体 {code, message, data}，HTTP 状态码即业务语义。
// 成功时把整个包体交给调用方（res.data 即业务数据），失败时 reject {status, message}。
export function request(config) {
    let vue = Vue.prototype;
    config.animation = config.animation ?? false
    config.dialog = config.dialog ?? true

    const instance = axios.create({
        // 开发环境由 vue.config.js 的 devServer 代理到后端，生产为同源，故根路径为空
        baseURL: "",
        timeout: 30 * 1000,
    })

    let loading;

    function startLoading() {
        loading = vue.loading.show()
    }

    function endLoading() {
        setTimeout(() => {
            loading && loading.close()
        }, 251)
    }

    instance.interceptors.request.use(request => {
        if (config.animation) {
            startLoading();
        }
        const token = localStorage.getItem('token')
        request.headers = {
            "Content-type": "application/json",
            "Authorization": token ? ('Bearer ' + token) : ''
        }
        return request
    }, err => {
        console.log(err)
    })

    let href = window.location.href
    instance.interceptors.response.use(res => {
        endLoading()
        return res.data
    }, err => {
        endLoading()
        let status = -1
        if (err.response) {
            status = err.response.status
        }
        const message = (err.response && err.response.data && err.response.data.message) || '服务器内部错误'

        if (status === -1) {
            if (config.dialog) {
                vue.dialog.show({
                    title: "请求失败",
                    content: '请求超时，请检查网络',
                    leftShow: false,
                    rightText: "重新加载",
                    type: "danger"
                }).onRightClick(() => {
                    setTimeout(() => {
                        window.location.reload()
                    }, 300)
                })
            }
            return Promise.reject({status, message: '请求超时'})
        }

        if (status === 401) {
            if (href.includes("/login")) {
                // 登录页的 401 即账号或密码错误
                vue.dialog.show({
                    title: "提示",
                    content: message || "登录失败，账号或密码错误",
                    leftShow: false,
                    type: "info"
                })
            } else {
                localStorage.removeItem('token')
                vue.dialog.show({
                    title: "登录失效",
                    content: "请重新登录",
                    leftShow: false,
                    autoClose: false,
                    type: "danger",
                }).onRightClick(dlg => {
                    dlg.close()
                    setTimeout(() => {
                        window.location.href = href.substring(0, href.indexOf("/")) + "/login"
                    })
                })
            }
            return Promise.reject({status, message})
        }

        if (config.dialog && sessionStorage.getItem("serverError") !== "true") {
            sessionStorage.setItem("serverError", "true")
            vue.dialog.show({
                title: "请求失败",
                content: message,
                leftShow: false,
                type: status >= 500 ? "danger" : "warn"
            }).onOutsideClick(() => {
                sessionStorage.setItem("serverError", "false")
            }).onRightClick(() => {
                sessionStorage.setItem("serverError", "false")
            })
        }
        return Promise.reject({status, message})
    })

    return instance(config)
}
