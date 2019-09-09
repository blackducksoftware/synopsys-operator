import { module, test } from 'qunit';
import { setupTest } from 'ember-qunit';

module('Unit | Route | documentation/eks', function(hooks) {
  setupTest(hooks);

  test('it exists', function(assert) {
    let route = this.owner.lookup('route:documentation/eks');
    assert.ok(route);
  });
});
