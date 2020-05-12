import { expect } from "chai";
import { shallowMount } from "@vue/test-utils";
import moxios from "moxios";
import Runs from "@/views/Runs.vue";
import RunItem from "@/components/RunItem.vue";

describe("Runs.vue", () => {
  beforeEach(() => {
    moxios.install();
  });

  afterEach(() => {
    moxios.uninstall();
  });

  it("displays a loading message while loading runs", () => {
    const wrapper = shallowMount(Runs);
    expect(wrapper.find("#runs-info").text()).to.contain("Loading");
  });

  it("displays an error message when the fetch fails", (done) => {
    moxios.stubRequest("/api/runs", { status: 500 });

    const wrapper = shallowMount(Runs);
    moxios.wait(() => {
      expect(wrapper.find("#runs-info").text()).to.contain("error");
      done();
    });
  });

  it("displays a run for each item in the runs list", (done) => {
    moxios.stubRequest("/api/runs", {
      status: 200,
      response: {
        items: [{}, {}, {}],
      },
    });

    const wrapper = shallowMount(Runs);
    moxios.wait(() => {
      expect(wrapper.findAll(RunItem).length).to.equal(3);
      done();
    });
  });
});
