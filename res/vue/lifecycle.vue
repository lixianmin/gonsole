<template>
  <div class="container">
    <div>click {{ count }} times</div>
    <button @click="clickBtn">更新</button>
  </div>
</template>

<script>
import { defineComponent } from '@vue/composition-api'

export default defineComponent({
  data() { return { count: 0 }},
  methods: { clickBtn:function() { this.count++ } },
  beforeCreate() {
    console.log('beforeCreate： 我在创建前显示');
  },
  created() {
    console.log('created： 我在创建后显示,可以在我的里面做一些数据请求');
  },
  beforeMount() {
    console.log('beforeMount： 我在挂载前显示');
  },
  mounted() {
    console.log('mounted： 我在挂载后显示，此时DOM已经被渲染出来了');
  },
  beforeUpdate() {
    console.log('\n');
    console.log('beforeUpdate： 我在更新前显示');
  },
  updated() {
    console.log('updated： 我在更新后显示');
  },
  beforeUnmount() {
    console.log('\n');
    console.log('beforeUnmount： 我在卸载前显示');
  },
  unmounted() {
    console.log('unmounted： 我在卸载后显示');
  },
})
</script>
