<template>
   <div>
      <v-toolbar>
         <v-spacer></v-spacer>

         <v-toolbar-items>
            <v-tooltip left>
               <v-btn
                    slot="activator" flat icon
                    v-on:click="fetchEnvs">
                  <v-icon color="green">sync</v-icon>
               </v-btn>

               <span>Reload environment list</span>
            </v-tooltip>
         </v-toolbar-items>
      </v-toolbar>

      <v-alert v-model="alert.show"
               dismissible
               transition="scale-transition"
               type="error">
         {{alert.text}}
      </v-alert>

      <v-data-table
           :headers="headers"
           :items="environments"
           :pagination.sync="pagination"
           no-data-text="No environments are currently running"
           class="elevation-1">

         <template slot="items" slot-scope="props">
            <td>{{props.item.id}}</td>
            <td>{{props.item.name}}</td>
            <td>{{props.item.description}}</td>
            <td>{{props.item.numOfContainers}}</td>
            <td>
               <ul>
                  <li v-for="port in props.item.ports">
                     {{port}}
                  </li>
               </ul>
            </td>
            <td>{{props.item.created}}</td>
            <td>{{props.item.keep_alive}}</td>
            <td class="text-xs-right">
               <v-tooltip bottom>
                  <v-btn
                       slot="activator"
                       flat icon v-on:click="envInfo(props.item)">
                     <v-icon color="indigo">info</v-icon>
                  </v-btn>

                  <span>Environment info</span>
               </v-tooltip>

               <v-tooltip bottom>
                  <v-btn
                       slot="activator"
                       flat icon v-on:click="deleteEnv(props.item.id)">
                     <v-icon color="pink">delete</v-icon>
                  </v-btn>

                  <span>Terminate environment</span>
               </v-tooltip>
            </td>
         </template>
      </v-data-table>

      <router-view></router-view>
   </div>

</template>

<script>
  export default {
    mounted() {
      this.fetchEnvs();
    },

    methods: {
      deleteEnv(id) {
        fetch(`/api/v1/env/${id}`, {method: "DELETE"})
          .then(this.fetchEnvs)
          .catch(error => {
            this.showError(`Error deleting environment: ${error.toString()}`)
          })
      },

      envInfo(env) {
        this.$router.push({name: 'env-info', params: {id: env.id}});
      },

      showError(msg) {
        this.alert.text = msg;
        this.alert.show = true;
      },

      clearError() {
        this.alert.text = "";
        this.alert.show = false;
      },

      fetchEnvs() {
        fetch('/api/v1/env')
          .then(stream => stream.json())
          .then(data => {
            this.clearError();
            this.environments = data.data.map(this.$root.$data.parseEnv);
          })
          .catch(error => {
            this.showError(`Error fetching environments: ${error.toString()}`);
          })
      }
    },

    data: () => ({
      alert: {
        show: false,
        text: ""
      },
      pagination: {
        rowsPerPage: -1,
        sortBy: "created",
        descending: true,
      },
      headers: [
        {
          text: 'Id',
          sortable: false,
          value: 'id'
        },
        {
          text: 'Name',
          sortable: true,
          value: 'name'
        },
        {
          text: 'Description',
          sortable: true,
          value: 'description'
        },
        {
          text: '# of containers',
          sortable: true,
          value: 'containers'
        },
        {
          text: 'Ports',
          sortable: false,
          value: 'ports'
        },
        {
          text: 'Created',
          sortable: true,
          value: 'created'
        },
        {
          text: 'Keepalive',
          sortable: true,
          value: 'keepalive'
        },
        {
          text: 'Actions',
          sortable: false,
          value: 'actions'
        },
      ],
      environments: []
    })
  }
</script>

<style>

</style>
