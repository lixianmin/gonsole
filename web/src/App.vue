<script setup lang="ts">
// This starter template is using Vue 3 <script setup> SFCs
// Check out https://v3.vuejs.org/api/sfc-script-setup.html#sfc-script-setup
import HelloWorld from './components/HelloWorld.vue'
import axios from 'axios'
import StartX from "./lib/starx";

class WebConfig {
  public constructor(config) {
    // console.log(config)
    this.autoLoginLimit = config.autoLoginLimit
    this.websocketPath = config.websocketPath
  }

  public getWebsocketUrl(): string {
    const isHttps = "https:" === document.location.protocol
    const protocol = isHttps ? "wss://" : "ws://"
    const url = `${protocol}${myHost}/${this.websocketPath}`
    return url
  }

  public readonly autoLoginLimit: number
  private readonly websocketPath: string
}

let myHost = "localhost:8888/ws"
const response = await axios.get(`${document.location.protocol}//${myHost}/web_config`)
const config = new WebConfig(response.data)

let star = new StartX()
star.init({url: config.getWebsocketUrl()}, () => {
  console.log("star initialized")
})

</script>

<template>
  <img alt="Vue logo" src="./assets/logo.png"/>
  <HelloWorld msg="Hello Vue 3 + TypeScript + Vite"/>
</template>

<style>
/*http://thomasf.github.io/solarized-css/*/
html {background-color: #002b36;color: #839496;margin: 1em;font-size: 1.2em;}
.copy_button { background-color: #008CBA; border: none; color: white; }

a {color: #b58900;}
a:visited {color: #cb4b16;}
a:hover {color: #cb4b16;}

table { border-width: 1px; border-color: #729ea5;border-collapse: collapse;}
th { background-color:#004949; border-width: 1px;padding: 8px;border-style: solid;border-color: #729ea5;text-align:left;}
th:hover { cursor: pointer;}
th:after { content: attr(data-text); font-size: small; margin-left: 5px;}
td { border-width: 1px;padding: 8px;border-style: solid;border-color: #729ea5;}

/*https://www.runoob.com/css/css-tooltip.html*/
.tips { position: relative; display: inline-block; border-bottom: 1px dotted black; }

.tips .tips_text {
  visibility: hidden; display: inline-block; white-space: nowrap; background: #005959; border-radius: 6px; padding: 6px 6px;
  /* 定位 */
  position: absolute; z-index: 1; top: -5px;left: 105%;
}

.tips:hover .tips_text { visibility: visible; }

#mainPanel {margin: 0;padding: 0.5em 0.5em 0.5em 0.5em;position: absolute;top: 0.5em;left: 0.5em;right: 0.5em;bottom: 3em;overflow: auto;}
#inputBoxDiv {padding: 0 0.5em 0 0.5em;margin: 0;position: absolute;bottom: 1em;left: 1px;width: 100%;overflow: hidden;}
#inputBox {width:100%;height:1.6em;font-size:1.5em; background-color: #073642; color: #859900}
</style>
