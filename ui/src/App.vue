<template>
  <div id="app" :class="theme">
    <AppHeader></AppHeader>
    <div class="view">
      <router-view />
    </div>
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
  min-width: 400px;
}

a {
  text-decoration: none;
  color: black;
}

.view {
  height: calc(100% - 40px);
  padding: 0 5px;
  overflow: auto;
}

:root {
  --pending-color: #c69162;
  --running-color: #6296c6;
  --successful-color: #67a567;
  --failed-color: #c67462;
  --canceled-color: #d2cdd5;
  --skipped-color: #d2cdd5;
}

.light {
  --box-background: white;
}
.light a {
  color: blue;
}

.dark {
  --box-background: #d4cede;
  background: #52475a;
  color: white;
}
.dark a {
  color: burlywood;
}
</style>
