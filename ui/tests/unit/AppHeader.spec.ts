import chai, { expect } from "chai";
import sinonChai from "sinon-chai";
import { shallowMount } from "@vue/test-utils";
import { spy } from "sinon";
import AppHeader from "@/components/AppHeader.vue";

chai.use(sinonChai);

describe("AppHeader.vue", () => {
  it("has the theme displayed according to the state", () => {
    const wrapper = shallowMount(AppHeader, {
      stubs: ["router-link"],
      mocks: {
        $store: {
          state: { theme: "light" },
        },
      },
    });
    expect(wrapper.find("#toggle-theme").text()).to.contain("light");
  });

  it("dispatches toggleTheme when clicking on the theme toggle", () => {
    const dispatch = spy();

    const wrapper = shallowMount(AppHeader, {
      stubs: ["router-link"],
      mocks: {
        $store: {
          state: { theme: "light" },
          dispatch,
        },
      },
    });

    wrapper.find("#toggle-theme").trigger("click");
    expect(dispatch).to.have.been.calledWith("toggleTheme");
  });
});
