/********************************************************************
 created:    2022-02-25
 author:     lixianmin

 Copyright (C) - All Rights Reserved
 *********************************************************************/
import {createApp} from 'vue'
import App from './App.vue'
import {createPinia} from "pinia";

createApp(App)
    .use(createPinia())
    .mount('#app')
