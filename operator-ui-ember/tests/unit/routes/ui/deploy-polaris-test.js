import { module, test } from 'qunit';
import { setupTest } from 'ember-qunit';

module('Unit | Route | ui/deploy_polaris', function(hooks) {
  setupTest(hooks);

  test('it exists', function(assert) {
    let route = this.owner.lookup('route:ui/deploy-polaris');
    assert.ok(route);
  });
});
