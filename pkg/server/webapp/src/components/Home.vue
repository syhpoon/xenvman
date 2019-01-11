<template>
   <div>
      <v-tabs v-model="active">

         <v-tab :key="envs" ripple>
            Environments
         </v-tab>

         <v-tab :key="tpls" ripple>
            Templates
         </v-tab>

         <!-- Environments list -->
         <v-tab-item :key="envs">
            <EnvBrowser/>
         </v-tab-item>

         <!-- Templates browser -->
         <v-tab-item :key="tpls">
            TPLS!
         </v-tab-item>

      </v-tabs>
   </div>
</template>

<script>
  import EnvBrowser from './EnvBrowser'

  export default {
    components: {
      EnvBrowser,
    },
    data: () => ({})
  }
</script>

<style>

</style>
