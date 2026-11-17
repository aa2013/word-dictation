<template>
  <div class="login-bg">
    <div class="login-container">
      <v-card elevation="2" class="login-card" :disabled="submitting || checking">
        <img src="../assets/img/logo.png" alt="logo" class="horizontal-center d-block" height="48" width="48">
        <p class="login-card-title m10t">系统初始化</p>
        <p class="setup-tip">检测到首次部署，请设置管理员账号</p>
        <v-form ref="setupForm">
          <v-text-field v-model="account" label="管理员账号" dense outlined
                        prepend-inner-icon="mdi-account"
                        :rules="[v=>(v&&v.trim().length>0)||'请输入管理员账号']"/>
          <v-text-field v-model="password" type="password" label="密码" dense outlined
                        prepend-inner-icon="mdi-lock"
                        :rules="[v=>(v&&/^(?=.*\d)(?=.*[a-z])(?=.*[A-Z]).+$/.test(v))||'必须包含大小写字母和数字',v=>(v&&v.length>=8&&v.length<=16)||'长度必须为8~16位']"/>
          <v-text-field v-model="passwordAgain" type="password" label="再次输入密码" dense outlined
                        prepend-inner-icon="mdi-lock"
                        :rules="[v=>(v&&v===password)||'两次密码必须相等']"/>
        </v-form>
        <v-btn block elevation="2" large rounded :loading="submitting"
               class="login-btn" @click="submit">
          完成初始化
        </v-btn>
        <div class="d-flex justify-center">
          <v-btn text class="forget-pwd" @click="gotoLogin">
            返回登录
          </v-btn>
        </div>
      </v-card>
    </div>
  </div>
</template>

<script>
import * as user from '@/network/details/user'

export default {
  name: "Setup",
  data: () => ({
    account: 'admin',
    password: '',
    passwordAgain: '',
    submitting: false,
    checking: true
  }),
  created() {
    user.setupStatus().then(res => {
      if (res.data && res.data.initialized) {
        this.$router.replace("/login")
      }
      this.checking = false
    }).catch(() => {
      this.checking = false
    })
  },
  methods: {
    gotoLogin() {
      this.$router.push("/login")
    },
    submit() {
      if (this.submitting || this.checking) {
        return
      }
      if (!this.$refs.setupForm.validate()) {
        return
      }
      this.submitting = true
      user.initSystem({
        account: this.account.trim(),
        password: this.password,
        passwordAgain: this.passwordAgain
      }).then(res => {
        if (res.data === true) {
          this.snackBar.show("初始化完成，请使用新账号登录")
          this.$router.replace("/login")
        }
      }).catch(() => {
      }).finally(() => {
        this.submitting = false
      })
    }
  }
}
</script>

<style scoped lang="less">

.setup-tip {
  text-align: center;
  font-size: 13px;
  color: #999;
  margin: -16px 0 16px;
}

.login-bg {
  background: url("../assets/img/bg.png");
  position: fixed;
  width: 100%;
  height: 100%;
  background-size: 100% 100%;

  .login-container {
    user-select: none;
    position: absolute;
    left: 50%;
    top: 50%;
    transform: translate(-50%, -50%);
    width: 320px;
    min-height: 400px;

    .login-card {
      width: 100%;
      padding: 10px 20px 5px;

      .login-card-title {
        margin-bottom: 24px;
        text-align: center;
        font-weight: bold;
        font-size: 24px;
        color: #1485FE;
      }

      .login-btn {
        margin: 10px 0;
        color: white;
        font-weight: bold;
        font-size: 16px;
        background-image: linear-gradient(to right, #3E50F7, #1485FE);
      }

      .forget-pwd {
        display: inline-block;
        line-height: 28px;
        text-align: center;
        font-size: 14px;
        color: #1485FE;
        cursor: pointer;
      }
    }
  }
}
</style>
