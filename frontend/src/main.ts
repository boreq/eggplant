import Vue from 'vue';
import App from './App.vue';
import router from './router';
import store from './store';
import VTooltip from 'v-tooltip';
import VueTimeago from 'vue-timeago';

Vue.use(VTooltip);

Vue.use(VueTimeago, {});

Vue.config.productionTip = false;

new Vue({
    router,
    store,
    render: (h) => h(App),
}).$mount('#app');

if ('serviceWorker' in navigator) {
    window.addEventListener('load', () => {
        navigator.serviceWorker.register('/sw.js').catch((err) => {
            console.error('service worker registration failed', err);
        });
    });
}
