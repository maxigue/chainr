<template>
  <div class="home">
    <h1>Welcome to Chainr!</h1>
    <p>
      This is Chainr's monitoring interface. It gives an overview of the runs in
      progress.
    </p>
    <p>
      For a detailed view of the runs, see the
      <router-link to="/runs">Runs</router-link> page.
    </p>
    <p>
      For more information about the project, see the
      <router-link to="/about">About</router-link> page.
    </p>
    <p id="nb-runs" v-if="loading">Loading runs...</p>
    <p id="nb-runs" v-if="error">An error occurred while loading runs.</p>
    <p id="nb-runs" v-if="!loading && !error">
      There are currently {{ nbRuns }} runs in progress.
    </p>
  </div>
</template>

<script lang="ts">
import { Component, Vue } from "vue-property-decorator";
import axios from "axios";

interface Run {
  status: string;
}

@Component
export default class Home extends Vue {
  private loading = true;
  private error = false;
  private nbRuns = 0;

  private created() {
    axios
      .get("/api/runs")
      .then((response) => {
        this.nbRuns = response.data.items.reduce(
          (nbRuns: number, item: Run) => {
            if (item.status == "PENDING" || item.status == "RUNNING") {
              return nbRuns + 1;
            }
            return nbRuns;
          },
          0
        );
      })
      .catch(() => {
        this.error = true;
      })
      .finally(() => {
        this.loading = false;
      });
  }
}
</script>

<style scoped>
#nb-runs {
  margin-top: 50px;
}
</style>
