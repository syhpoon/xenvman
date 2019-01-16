<template>
   <div>
      <v-toolbar>
         <v-spacer></v-spacer>

         <v-toolbar-items>
            <v-tooltip left>
               <v-btn
                    slot="activator" flat icon
                    v-on:click="fetchTpls">
                  <v-icon color="green">sync</v-icon>
               </v-btn>

               <span>Reload templates list</span>
            </v-tooltip>
         </v-toolbar-items>
      </v-toolbar>

      <v-alert v-model="alert.show"
               dismissible
               transition="scale-transition"
               type="error">
         {{alert.text}}
      </v-alert>

      <v-layout justify-space-between pa-3>
         <v-flex xs-5>
            <v-treeview
                 :open="open"
                 :items="tpls"
                 :active.sync="active"
                 activatable
                 open-on-click>

               <template slot="prepend" slot-scope="{ item, open }">
                  <v-icon v-if="!item.file">
                     {{ open ? 'mdi-folder-open' : 'mdi-folder' }}
                  </v-icon>
                  <v-icon v-else>
                     {{ files[item.file] }}
                  </v-icon>
               </template>

            </v-treeview>
         </v-flex>

         <v-flex d-flex>
            <v-card v-if="selected">
               <div class="table-responsive">
                  <table class="info-table">
                     <tbody>
                     <tr>
                        <td>Name</td>
                        <td>{{selected.id}}</td>
                     </tr>

                     <tr>
                        <td>Description</td>
                        <td>{{selected.tpl.description}}</td>
                     </tr>

                     <tr v-if="selected.tpl.parameters">
                        <td>Parameters</td>
                        <td>
                           <table class="info-table">
                              <thead>
                              <tr>
                                 <th>Name</th>
                                 <th>Description</th>
                                 <th>Type</th>
                              </tr>
                              </thead>

                              <tbody>

                              <tr v-for="(p, k) in selected.tpl.parameters">
                                 <td>{{k}}</td>
                                 <td>{{p.description}}</td>
                                 <td>{{p.type}}</td>
                              </tr>

                              </tbody>
                           </table>
                        </td>
                     </tr>

                     <tr v-if="selected.tpl.data_dir.length > 0">
                        <td>Data dir files</td>
                        <td>
                           <ul>
                              <li v-for="f in selected.tpl.data_dir">
                                 {{f}}
                              </li>
                           </ul>
                        </td>
                     </tr>
                     </tbody>
                  </table>
               </div>
            </v-card>
         </v-flex>
      </v-layout>
   </div>

</template>

<script>
  export default {
    mounted() {
      this.fetchTpls();
    },

    computed: {
      selected() {
        if (!this.active.length) {
          return undefined;
        }

        const id = this.active[0];

        return this.all_tpls.find(tpl => tpl.id === id);
      },
    },

    methods: {
      showError(msg) {
        this.alert.text = msg;
        this.alert.show = true;
      },

      clearError() {
        this.alert.text = "";
        this.alert.show = false;
      },

      fetchTpls() {
        fetch('/api/v1/tpl')
          .then(stream => stream.json())
          .then(data => {
            this.clearError();
            const d = data.data;
            let tpls = [];
            let cur = tpls;
            let prefix = [];

            // Populate dir hierarchy
            Object.keys(d).forEach(tplName => {
              let path = tplName.split("/");
              let slice = path.slice(0, -1);

              for (let i = 0; i < slice.length; i++) {
                let idx = tpls.findIndex(el => {
                  return el.name === slice[i] && el.type === 'dir';
                });

                if (idx === -1) {
                  cur.push({
                             name: slice[i],
                             id: slice[i],
                             type: 'dir',
                             children: []
                  });

                  cur = cur[cur.length - 1].children;
                } else {
                  cur = tpls[idx].children;
                }

                prefix.push(slice[i]);
              }

              let name = path[path.length - 1];

              prefix.push(name);

              let id = prefix.join("/");

              cur.push({
                         id: id,
                         name: name,
                         file: 'tpl'
                       });

              this.all_tpls.push(
                {
                  id: id,
                  tpl: d[tplName]
                });

              cur = tpls;
              prefix = [];
            });

            this.tpls = tpls;
          })
          .catch(error => {
            this.showError(`Error fetching templates: ${error.toString()}`);
          })
      }
    },

    data: () => ({
      alert: {
        show: false,
        text: ""
      },
      files: {
        tpl: 'mdi-language-javascript'
      },
      open: [],
      tpls: [],
      all_tpls: [],
      active: []
    })
  }
</script>

<style>
</style>
