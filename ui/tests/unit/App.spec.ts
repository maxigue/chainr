import { expect } from "chai";
import { shallowMount } from "@vue/test-utils";
import App from "@/App.vue";

describe("App.vue", () => {
  it("has class corresponding to the theme", () => {
    const wrapper = shallowMount(App, {
      stubs: ["router-view"],
      mocks: {
        $store: {
          state: { theme: "light" },
        },
      },
    });
    expect(wrapper.classes()).to.contain("light");
  });
});
