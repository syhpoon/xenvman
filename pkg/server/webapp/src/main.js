import Vue from 'vue'
import VueRouter from 'vue-router'

import App from './App.vue'
import EnvInfo from './components/EnvInfo'
import Home from './components/Home'

import './plugins/vuetify'

Vue.config.productionTip = false;

Vue.use(VueRouter);

const routes = [
	{
		path: '/',
		name: 'home',
		component: Home,
		children: [
			{
				path: 'env/info/:id',
				name: 'env-info',
				component: EnvInfo
			}
		]
	},
];

const router = new VueRouter({routes});
const lib = {
	parseEnv(env) {
		env.numOfContainers = 0;
		env.ports = [];
		env.containers = [];

		for (const tplName in env.templates) {
			const tpl = env.templates[tplName];

			tpl.forEach(t => {
				const keys = Object.keys(t.containers);

				env.numOfContainers += keys.length;

				keys.forEach(contKey => {
					const cont = t.containers[contKey];
					const pKeys = Object.keys(cont.ports);

					let c = {"hostname": cont.hostname, "ports": []};

					pKeys.forEach(iport => {
						env.ports.push(`${iport} -> ${cont.ports[iport]}`);
						c.ports.push(`${iport} -> ${cont.ports[iport]}`);
					});

					env.containers.push(c);
				});
			});
		}

		return env;
	}
};

new Vue({
	router,
	data: lib,
	render: h => h(App),
}).$mount('#app');
