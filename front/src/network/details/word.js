import {request} from "../request";

// 导入单个单词
export function importSingle({libId, word}) {
    return request({
        url: `/api/libs/${libId}/words`,
        method: 'post',
        data: {word},
        animation: false,
        dialog: false
    })
}

export function getWordList({libId, word, pageNum, pageSize}) {
    return request({
        url: `/api/libs/${libId}/words`,
        method: 'get',
        params: {q: word, pageNum, pageSize},
        animation: false
    })
}

export function getRandomWordList({libId, size}) {
    return request({
        url: `/api/libs/${libId}/random-words`,
        method: 'get',
        params: {size},
    })
}

export function getExplains({wordId, libId}) {
    return request({
        url: `/api/words/${wordId}/explains`,
        method: 'get',
        params: {libId},
    })
}

export function changeDefaultExplain({libId, wordId, defaultId}) {
    return request({
        url: `/api/libs/${libId}/default-explain`,
        method: 'put',
        data: {wordId, explainId: defaultId},
    })
}
