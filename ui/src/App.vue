<template>
  <div id="app" :class="theme">
    <AppHeader></AppHeader>
    <router-view />
  </div>
</template>

<script lang="ts">
import { Component, Vue } from "vue-property-decorator";
import AppHeader from "@/components/AppHeader.vue";

@Component({
  components: {
    AppHeader,
  },
})
export default class App extends Vue {
  private get theme() {
    return this.$store.state.theme;
  }
}
</script>

<style scoped>
#app {
  height: 100%;
  font-family: Avenir, Helvetica, Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  text-align: center;
}
</style>

<style>
html,
body {
  height: 100%;
  margin: 0;
  padding: 0;
}

a {
  text-decoration: none;
  color: black;
}

.light {
}
.light a {
  color: blue;
}

.dark {
  background: #52475a;
  color: white;
}
.dark a {
  color: burlywood;
}
</style>
