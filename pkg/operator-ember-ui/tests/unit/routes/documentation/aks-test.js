import { module, test } from 'qunit';
import { setupTest } from 'ember-qunit';

module('Unit | Route | documentation/aks', function(hooks) {
  setupTest(hooks);

  test('it exists', function(assert) {
    let route = this.owner.lookup('route:documentation/aks');
    assert.ok(route);
  });
});
