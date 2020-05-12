<template>
  <div class="run">
    <div class="status-indicator" :class="run.status.toLowerCase()"></div>
    <div class="content">
      <div class="status" :class="run.status.toLowerCase()">
        {{ run.status }}
      </div>
      <div class="progress-container">
        <div class="progress-bar">
          <div
            v-for="(job, index) of run.jobs"
            :key="index"
            class="job"
            :class="job.status.toLowerCase()"
          >
            {{ job.name }}
          </div>
        </div>
      </div>
    </div>
  </div>
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

<style scoped>
.run {
  display: flex;
  margin-top: 5px;
  height: 100px;
  border: 1px solid black;
  border-radius: 25px;
  overflow: hidden;
  background: var(--box-background);
}

.status-indicator {
  width: 20px;
}
.status-indicator.pending {
  background: var(--pending-color);
}
.status-indicator.running {
  background: var(--running-color);
}
.status-indicator.successful {
  background: var(--successful-color);
}
.status-indicator.failed {
  background: var(--failed-color);
}
.status-indicator.canceled {
  background: var(--canceled-color);
}

.content {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  padding: 5px;
}

.status {
  height: 20px;
  font-weight: bold;
}
.status.pending {
  color: var(--pending-color);
}
.status.running {
  color: var(--running-color);
}
.status.successful {
  color: var(--successful-color);
}
.status.failed {
  color: var(--failed-color);
}
.status.canceled {
  color: var(--canceled-color);
}

.progress-container {
  height: 100%;
  width: 100%;
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
}

.progress-bar {
  height: 30px;
  width: 80%;
  display: flex;
  border: 1px solid black;
  border-radius: 25px;
  overflow: hidden;
}

.job {
  flex: 1;
  display: flex;
  justify-content: center;
  align-items: center;
  border-right: 1px solid black;
}
.job:last-child {
  border-right: none;
}
.job.pending {
  background: var(--pending-color);
}
.job.skipped {
  background: var(--skipped-color);
}
.job.running {
  background: var(--running-color);
}
.job.successful {
  background: var(--successful-color);
}
.job.failed {
  background: var(--failed-color);
}
</style>
