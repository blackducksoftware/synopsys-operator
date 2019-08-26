import { module, test } from 'qunit';
import { setupTest } from 'ember-qunit';

module('Unit | Route | ui/operator', function(hooks) {
  setupTest(hooks);

  test('it exists', function(assert) {
    let route = this.owner.lookup('route:ui/operator');
    assert.ok(route);
  });
});
