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
    <br />
    <br />
    <p id="nb-runs" v-if="loading">Loading runs{{ suspensionPoints }}</p>
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
  private suspensionPoints = ".";

  private loading = true;
  private error = false;
  private nbRuns = 0;

  private created() {
    const interval = setInterval(() => {
      this.suspensionPoints = this.updateSuspensionPoints(
        this.suspensionPoints
      );
    }, 500);

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
        clearInterval(interval);
      });
  }

  private updateSuspensionPoints(suspensionPoints: string): string {
    let next = suspensionPoints + ".";
    if (next.length > 3) {
      next = ".";
    }
    return next;
  }
}
</script>
