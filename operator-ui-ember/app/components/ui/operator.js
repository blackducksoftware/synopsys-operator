import Component from '@ember/component';

export default Component.extend({
  operatorConfig: {
    namespace: "",
    clusterScoped: false,
    enableAlert: false,
    enableBlackDuck: false,
    enableOpsSight: false,
    enablePolaris: false,
    exposeMetrics: "",
    exposeUI: "",
    metricsImage: "",
    operatorImage: ""
  },
  actions: {
    deployOperator() {
      $.ajax({
        type: "POST",
        url: "/api/deploy_operator",
        data: JSON.stringify(this.operatorConfig),
        success: function () {
          alert("success")
        }
      });
    }
  }
});
