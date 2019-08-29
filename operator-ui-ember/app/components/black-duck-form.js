import Component from '@ember/component';

export default Component.extend({
    deployed: false,
    actions: {
        deploy() {
            alert("deploying")
            alert(this.model)
            this.toggleProperty('deployed')
        }
    }
});
