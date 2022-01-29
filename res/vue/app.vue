<template>
  <div>
    <Lifecycle v-if="show"></Lifecycle>
    <button @click="unMount">卸载</button>
  </div>
</template>

<script>
import Lifecycle from '@/lifecycle.vue'

export default {
  name: 'App',
  components: {
    Lifecycle: Lifecycle,
  },
  data() {return { show: true }},
  methods: {
    unMount:function() { this.show = !this.show; }
  }
}
</script>
