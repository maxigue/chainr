import Vue from "vue";
import Vuex from "vuex";

Vue.use(Vuex);

export default new Vuex.Store({
  state: {
    theme: "light",
  },
  mutations: {
    toggleTheme(state) {
      if (state.theme == "light") {
        state.theme = "dark";
      } else {
        state.theme = "light";
      }
    },
  },
  actions: {
    toggleTheme({ commit }) {
      commit("toggleTheme");
    },
  },
  modules: {},
});
