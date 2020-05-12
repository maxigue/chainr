<template>
  <div class="runs">
    <h1>Runs</h1>
    <p id="runs-info" v-if="loading">Loading runs...</p>
    <p id="runs-info" v-else-if="error">
      An error occurred while loading runs.
    </p>
    <p id="runs-info" v-else-if="runs.length == 0">
      There are no runs yet.<br />
      You can create a run by using the API.<br />
      For more information, you can read the documentation on the
      <a href="https://github.com/Tyrame/chainr" target="_blank">Github page</a>
    </p>
    <div id="runs" v-else>
      <RunItem v-for="(run, index) of runs" :key="index" :run="run"></RunItem>
    </div>
  </div>
</template>

<script lang="ts">
import { Component, Vue } from "vue-property-decorator";
import axios from "axios";
import RunItem from "@/components/RunItem.vue";

@Component({
  components: {
    RunItem,
  },
})
export default class Runs extends Vue {
  private loading = true;
  private error = false;
  private runs = [];
  private stop = false;

  private created() {
    this.updateRuns().finally(() => {
      this.loading = false;
    });
  }

  private destroyed() {
    this.stop = true;
  }

  private updateRuns() {
    return axios
      .get("/api/runs")
      .then((response) => {
        this.runs = response.data.items;
        this.error = false;
      })
      .catch(() => {
        this.error = true;
      })
      .finally(() => {
        setTimeout(() => {
          if (!this.stop) {
            this.updateRuns();
          }
        }, 2000);
      });
  }
}
</script>

<style scoped>
#runs {
  margin: 0 10%;
  overflow: auto;
  padding-bottom: 5px;
}
</style>
