import { expect } from "chai";
import { shallowMount } from "@vue/test-utils";
import moxios from "moxios";
import Home from "@/views/Home.vue";

describe("Home.vue", () => {
  beforeEach(() => {
    moxios.install();
  });

  afterEach(() => {
    moxios.uninstall();
  });

  it("displays a loading message while loading runs", () => {
    const wrapper = shallowMount(Home, {
      stubs: ["router-link"],
    });
    expect(wrapper.find("#nb-runs").text()).to.contain("Loading");
  });

  it("displays the number of runs with status RUNNING or PENDING when successfully fetched", (done) => {
    moxios.stubRequest("/api/runs", {
      status: 200,
      response: {
        items: [
          { status: "RUNNING" },
          { status: "PENDING" },
          { status: "SUCCESSFUL" },
          { status: "FAILED" },
        ],
      },
    });

    const wrapper = shallowMount(Home, {
      stubs: ["router-link"],
    });
    moxios.wait(() => {
      expect(wrapper.find("#nb-runs").text()).to.contain("2");
      done();
    });
  });

  it("Displays an error message when the fetch fails", (done) => {
    moxios.stubRequest("/api/runs", {
      status: 500,
    });

    const wrapper = shallowMount(Home, {
      stubs: ["router-link"],
    });
    moxios.wait(() => {
      expect(wrapper.find("#nb-runs").text()).to.contain("error");
      done();
    });
  });
});
