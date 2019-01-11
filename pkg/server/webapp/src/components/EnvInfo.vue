<template>
   <v-dialog v-model="show" width="60%" persistent>
      <v-toolbar>
         <b>Environment info</b>
         <v-spacer></v-spacer>

         <v-toolbar-items>
            <v-tooltip bottom>
               <v-btn slot="activator" flat icon
                      v-on:click="close">
                  <v-icon color="black">close</v-icon>
               </v-btn>

               <span>Close</span>
            </v-tooltip>
         </v-toolbar-items>
      </v-toolbar>

      <v-card>
         <div class="table-responsive">
            <table id="env-info-table">
               <tbody>
               <tr>
                  <td>Id</td>
                  <td>{{env.id}}</td>
               </tr>

               <tr>
                  <td>Name</td>
                  <td>{{env.name}}</td>
               </tr>

               <tr>
                  <td>Description</td>
                  <td>{{env.description}}</td>
               </tr>

               <tr>
                  <td>Workspace dir</td>
                  <td>{{env.ws_dir}}</td>
               </tr>
               <tr>
                  <td>Mount dir</td>
                  <td>{{env.mount_dir}}</td>
               </tr>
               <tr>
                  <td>Net ID</td>
                  <td>{{env.net_id}}</td>
               </tr>
               <tr>
                  <td>Created</td>
                  <td>{{env.created}}</td>
               </tr>
               <tr>
                  <td>Keep alive</td>
                  <td>{{env.keep_alive}}</td>
               </tr>
               <tr>
                  <td>External address</td>
                  <td>{{env.external_address}}</td>
               </tr>
               <tr>
                  <td>Containers</td>
                  <td>
                     <ul>
                        <li v-for="cont in env.containers">
                           {{cont.hostname}}

                           <ul>
                              <li v-for="port in cont.ports">
                                 {{port}}
                              </li>
                           </ul>
                        </li>
                     </ul>
                  </td>
               </tr>
               </tbody>
            </table>
         </div>
      </v-card>

   </v-dialog>
</template>

<script>
  export default {
    mounted() {
      const id = this.$route.params.id;

      //TODO
      fetch(`http://localhost:9876/api/v1/env/${id}`)
        .then(stream => stream.json())
        .then(data => {
          this.env = this.$root.$data.parseEnv(data.data);
        })
        .catch(error => {
          //TODO
        });
    },

    methods: {
      close() {
        this.show = false;
        this.$router.push({name: "home"});
      },
    },

    data: () => ({
      show: true,
      env: {}
    })
  }
</script>

<style>
   #env-info-table {
      font-family: "Roboto", Arial, Helvetica, sans-serif;
      border-collapse: collapse;
      width: 100%;
   }

   #env-info-table td {
      border: 1px solid #ddd;
      padding: 8px;
   }

   #env-info-table tr:nth-child(even) {
      background-color: #f2f2f2;
   }

   #env-info-table tr:hover {
      background-color: #ddd;
   }

   #customers td {
      padding-top: 12px;
      padding-bottom: 12px;
      text-align: left;
      background-color: #4CAF50;
      color: white;
   }
</style>
