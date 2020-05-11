<template>
  <div class="run">RUN {{ run.status }}</div>
</template>

<script lang="ts">
import { Component, Vue, Prop } from "vue-property-decorator";

interface Run {
  status: string;
  jobs: Array<Job>;
}

interface Job {
  name: string;
  status: string;
}

@Component
export default class RunItem extends Vue {
  @Prop({ required: true }) run!: Run;
}
</script>
