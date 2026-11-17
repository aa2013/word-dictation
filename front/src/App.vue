<template>
    <v-app style="width: 100%;height: 100%">
        <router-view/>
    </v-app>
</template>

<script>

export default {
    name: 'app',
    created() {
        // 判断是否登录，未登录则跳转登录页面（初始化页面除外）
        if (!localStorage.getItem('token')) {
            const path = window.location.pathname
            if (path !== "/login" && path !== "/setup") {
                this.$router.push("/login")
            }
        }
    },
    mounted() {

    }
}
</script>

<style>
</style>
