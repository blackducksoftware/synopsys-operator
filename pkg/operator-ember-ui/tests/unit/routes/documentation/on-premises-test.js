import { module, test } from 'qunit';
import { setupTest } from 'ember-qunit';

module('Unit | Route | documentation/on-premises', function(hooks) {
  setupTest(hooks);

  test('it exists', function(assert) {
    let route = this.owner.lookup('route:documentation/on-premises');
    assert.ok(route);
  });
});
