<template>
  <div id="header">
    <div class="home">
      <router-link to="/"><img src="@/assets/logo.png" />Chainr</router-link>
    </div>
    <div class="nav">
      <router-link to="/runs">Runs</router-link> |
      <router-link to="/about">About</router-link>
    </div>
    <a id="toggle-theme" @click="toggleTheme">Theme: {{ theme }}</a>
  </div>
</template>

<script lang="ts">
import { Component, Vue } from "vue-property-decorator";

@Component
export default class AppHeader extends Vue {
  private get theme() {
    return this.$store.state.theme;
  }

  private toggleTheme() {
    this.$store.dispatch("toggleTheme");
  }
}
</script>

<style scoped>
#header {
  height: 30px;
  padding: 5px;
  background: #ae8cc4;
  display: flex;
  justify-content: space-between;
  align-items: center;
  color: black;
  cursor: default;
}

#header a {
  color: black;
}
a:hover {
  color: white;
}

.home {
  font-size: 2em;
  font-weight: bold;
}
.home a {
  display: flex;
  align-items: center;
}
.home a:hover {
  color: black;
}

.nav a.router-link-active {
  font-weight: bold;
}
.nav a.router-link-active:hover {
  color: black;
  cursor: default;
}

#toggle-theme {
  cursor: pointer;
}
</style>
