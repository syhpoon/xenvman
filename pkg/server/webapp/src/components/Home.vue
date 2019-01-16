<template>
   <div>
      <v-tabs>

         <v-tab ripple>
            Environments
         </v-tab>

         <v-tab ripple>
            Templates
         </v-tab>

         <!-- Environments list -->
         <v-tab-item>
            <EnvBrowser/>
         </v-tab-item>

         <!-- Templates browser -->
         <v-tab-item>
            <TplBrowser/>
         </v-tab-item>

      </v-tabs>
   </div>
</template>

<script>
  import EnvBrowser from './EnvBrowser'
  import TplBrowser from './TplBrowser'

  export default {
    components: {
      EnvBrowser,
      TplBrowser,
    },
    data: () => ({})
  }
</script>

<style>
</style>
