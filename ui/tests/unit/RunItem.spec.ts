import { expect } from "chai";
import { shallowMount } from "@vue/test-utils";
import RunItem from "@/components/RunItem.vue";

describe("Runs.vue", () => {
  const run = {
    status: "RUNNING",
    jobs: [
      {
        name: "job1",
        status: "PENDING",
      },
      {
        name: "job2",
        status: "SUCCESSFUL",
      },
    ],
  };

  it("displays an indicator with class corresponding to the run status", () => {
    const wrapper = shallowMount(RunItem, {
      propsData: { run },
    });
    expect(wrapper.find(".status-indicator").classes()).to.contain("running");
  });

  it("displays the run status with the class corresponding to the run status", () => {
    const wrapper = shallowMount(RunItem, {
      propsData: { run },
    });
    expect(wrapper.find(".status").classes()).to.contain("running");
    expect(wrapper.find(".status").text()).to.equal("RUNNING");
  });

  it("has a div for each job in the progress bar", () => {
    const wrapper = shallowMount(RunItem, {
      propsData: { run },
    });
    expect(wrapper.find(".progress-bar").findAll(".job").length).to.equal(2);
  });

  it("adds the class corresponding to the job status on each job", () => {
    const wrapper = shallowMount(RunItem, {
      propsData: { run },
    });
    const jobs = wrapper.find(".progress-bar").findAll(".job");
    expect(jobs.at(0).classes()).to.contain("pending");
    expect(jobs.at(1).classes()).to.contain("successful");
  });

  it("displays the job name on each job", () => {
    const wrapper = shallowMount(RunItem, {
      propsData: { run },
    });
    const jobs = wrapper.find(".progress-bar").findAll(".job");
    expect(jobs.at(0).text()).to.equal("job1");
    expect(jobs.at(1).text()).to.equal("job2");
  });
});
