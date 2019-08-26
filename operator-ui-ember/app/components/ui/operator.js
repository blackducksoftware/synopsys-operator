import Component from '@ember/component';

export default Component.extend({
  didInsertElement() {
    fetch("http://10.145.119.53:8080/operator/status").then(
      function(response){
        return response.json();
      }
    ).then(
      data => {
        this.set('operator', data);
      }
    );
  }
});
