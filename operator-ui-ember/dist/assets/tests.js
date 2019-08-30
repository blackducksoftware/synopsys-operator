'use strict';

define("operator-docs/tests/integration/components/black-duck-form-test", ["qunit", "ember-qunit", "@ember/test-helpers"], function (_qunit, _emberQunit, _testHelpers) {
  "use strict";

  (0, _qunit.module)('Integration | Component | black-duck-form', function (hooks) {
    (0, _emberQunit.setupRenderingTest)(hooks);
    (0, _qunit.test)('it renders', async function (assert) {
      // Set any properties with this.set('myProperty', 'value');
      // Handle any actions with this.set('myAction', function(val) { ... });
      await (0, _testHelpers.render)(Ember.HTMLBars.template({
        "id": "kxmRmKXO",
        "block": "{\"symbols\":[],\"statements\":[[5,\"black-duck-form\",[],[[],[]]]],\"hasEval\":false}",
        "meta": {}
      }));
      assert.equal(this.element.textContent.trim(), ''); // Template block usage:

      await (0, _testHelpers.render)(Ember.HTMLBars.template({
        "id": "ZC40/S8o",
        "block": "{\"symbols\":[],\"statements\":[[0,\"\\n      \"],[5,\"black-duck-form\",[],[[],[]],{\"statements\":[[0,\"\\n        template block text\\n      \"]],\"parameters\":[]}],[0,\"\\n    \"]],\"hasEval\":false}",
        "meta": {}
      }));
      assert.equal(this.element.textContent.trim(), 'template block text');
    });
  });
});
define("operator-docs/tests/integration/components/documentation-navbar-test", ["qunit", "ember-qunit", "@ember/test-helpers"], function (_qunit, _emberQunit, _testHelpers) {
  "use strict";

  (0, _qunit.module)('Integration | Component | documentation-navbar', function (hooks) {
    (0, _emberQunit.setupRenderingTest)(hooks);
    (0, _qunit.test)('it renders', async function (assert) {
      // Set any properties with this.set('myProperty', 'value');
      // Handle any actions with this.set('myAction', function(val) { ... });
      await (0, _testHelpers.render)(Ember.HTMLBars.template({
        "id": "nGR0ZFgp",
        "block": "{\"symbols\":[],\"statements\":[[5,\"documentation-navbar\",[],[[],[]]]],\"hasEval\":false}",
        "meta": {}
      }));
      assert.equal(this.element.textContent.trim(), ''); // Template block usage:

      await (0, _testHelpers.render)(Ember.HTMLBars.template({
        "id": "O75PdVbo",
        "block": "{\"symbols\":[],\"statements\":[[0,\"\\n      \"],[5,\"documentation-navbar\",[],[[],[]],{\"statements\":[[0,\"\\n        template block text\\n      \"]],\"parameters\":[]}],[0,\"\\n    \"]],\"hasEval\":false}",
        "meta": {}
      }));
      assert.equal(this.element.textContent.trim(), 'template block text');
    });
  });
});
define("operator-docs/tests/integration/components/documentation/aks-test", ["qunit", "ember-qunit", "@ember/test-helpers"], function (_qunit, _emberQunit, _testHelpers) {
  "use strict";

  (0, _qunit.module)('Integration | Component | documentation/aks', function (hooks) {
    (0, _emberQunit.setupRenderingTest)(hooks);
    (0, _qunit.test)('it renders', async function (assert) {
      // Set any properties with this.set('myProperty', 'value');
      // Handle any actions with this.set('myAction', function(val) { ... });
      await (0, _testHelpers.render)(Ember.HTMLBars.template({
        "id": "290Mut7z",
        "block": "{\"symbols\":[],\"statements\":[[1,[23,\"documentation/aks\"],false]],\"hasEval\":false}",
        "meta": {}
      }));
      assert.equal(this.element.textContent.trim(), ''); // Template block usage:

      await (0, _testHelpers.render)(Ember.HTMLBars.template({
        "id": "oQAvVDvl",
        "block": "{\"symbols\":[],\"statements\":[[0,\"\\n\"],[4,\"documentation/aks\",null,null,{\"statements\":[[0,\"        template block text\\n\"]],\"parameters\":[]},null],[0,\"    \"]],\"hasEval\":false}",
        "meta": {}
      }));
      assert.equal(this.element.textContent.trim(), 'template block text');
    });
  });
});
define("operator-docs/tests/integration/components/documentation/deploy-operator-test", ["qunit", "ember-qunit", "@ember/test-helpers"], function (_qunit, _emberQunit, _testHelpers) {
  "use strict";

  (0, _qunit.module)('Integration | Component | documentation/deploy-operator', function (hooks) {
    (0, _emberQunit.setupRenderingTest)(hooks);
    (0, _qunit.test)('it renders', async function (assert) {
      // Set any properties with this.set('myProperty', 'value');
      // Handle any actions with this.set('myAction', function(val) { ... });
      await (0, _testHelpers.render)(Ember.HTMLBars.template({
        "id": "j3eOmiqO",
        "block": "{\"symbols\":[],\"statements\":[[1,[23,\"documentation/deploy-operator\"],false]],\"hasEval\":false}",
        "meta": {}
      }));
      assert.equal(this.element.textContent.trim(), ''); // Template block usage:

      await (0, _testHelpers.render)(Ember.HTMLBars.template({
        "id": "cCqJ59DL",
        "block": "{\"symbols\":[],\"statements\":[[0,\"\\n\"],[4,\"documentation/deploy-operator\",null,null,{\"statements\":[[0,\"        template block text\\n\"]],\"parameters\":[]},null],[0,\"    \"]],\"hasEval\":false}",
        "meta": {}
      }));
      assert.equal(this.element.textContent.trim(), 'template block text');
    });
  });
});
define("operator-docs/tests/integration/components/documentation/eks-test", ["qunit", "ember-qunit", "@ember/test-helpers"], function (_qunit, _emberQunit, _testHelpers) {
  "use strict";

  (0, _qunit.module)('Integration | Component | documentation/eks', function (hooks) {
    (0, _emberQunit.setupRenderingTest)(hooks);
    (0, _qunit.test)('it renders', async function (assert) {
      // Set any properties with this.set('myProperty', 'value');
      // Handle any actions with this.set('myAction', function(val) { ... });
      await (0, _testHelpers.render)(Ember.HTMLBars.template({
        "id": "B5b2+o4J",
        "block": "{\"symbols\":[],\"statements\":[[1,[23,\"documentation/eks\"],false]],\"hasEval\":false}",
        "meta": {}
      }));
      assert.equal(this.element.textContent.trim(), ''); // Template block usage:

      await (0, _testHelpers.render)(Ember.HTMLBars.template({
        "id": "1+cpGmTT",
        "block": "{\"symbols\":[],\"statements\":[[0,\"\\n\"],[4,\"documentation/eks\",null,null,{\"statements\":[[0,\"        template block text\\n\"]],\"parameters\":[]},null],[0,\"    \"]],\"hasEval\":false}",
        "meta": {}
      }));
      assert.equal(this.element.textContent.trim(), 'template block text');
    });
  });
});
define("operator-docs/tests/integration/components/documentation/gke-test", ["qunit", "ember-qunit", "@ember/test-helpers"], function (_qunit, _emberQunit, _testHelpers) {
  "use strict";

  (0, _qunit.module)('Integration | Component | documentation/gke', function (hooks) {
    (0, _emberQunit.setupRenderingTest)(hooks);
    (0, _qunit.test)('it renders', async function (assert) {
      // Set any properties with this.set('myProperty', 'value');
      // Handle any actions with this.set('myAction', function(val) { ... });
      await (0, _testHelpers.render)(Ember.HTMLBars.template({
        "id": "PTFA04VM",
        "block": "{\"symbols\":[],\"statements\":[[1,[23,\"documentation/gke\"],false]],\"hasEval\":false}",
        "meta": {}
      }));
      assert.equal(this.element.textContent.trim(), ''); // Template block usage:

      await (0, _testHelpers.render)(Ember.HTMLBars.template({
        "id": "2OZa/Bxq",
        "block": "{\"symbols\":[],\"statements\":[[0,\"\\n\"],[4,\"documentation/gke\",null,null,{\"statements\":[[0,\"        template block text\\n\"]],\"parameters\":[]},null],[0,\"    \"]],\"hasEval\":false}",
        "meta": {}
      }));
      assert.equal(this.element.textContent.trim(), 'template block text');
    });
  });
});
define("operator-docs/tests/integration/components/documentation/home-test", ["qunit", "ember-qunit", "@ember/test-helpers"], function (_qunit, _emberQunit, _testHelpers) {
  "use strict";

  (0, _qunit.module)('Integration | Component | documentation/home', function (hooks) {
    (0, _emberQunit.setupRenderingTest)(hooks);
    (0, _qunit.test)('it renders', async function (assert) {
      // Set any properties with this.set('myProperty', 'value');
      // Handle any actions with this.set('myAction', function(val) { ... });
      await (0, _testHelpers.render)(Ember.HTMLBars.template({
        "id": "N4261sB0",
        "block": "{\"symbols\":[],\"statements\":[[1,[23,\"documentation/home\"],false]],\"hasEval\":false}",
        "meta": {}
      }));
      assert.equal(this.element.textContent.trim(), ''); // Template block usage:

      await (0, _testHelpers.render)(Ember.HTMLBars.template({
        "id": "5m5Jk1lq",
        "block": "{\"symbols\":[],\"statements\":[[0,\"\\n\"],[4,\"documentation/home\",null,null,{\"statements\":[[0,\"        template block text\\n\"]],\"parameters\":[]},null],[0,\"    \"]],\"hasEval\":false}",
        "meta": {}
      }));
      assert.equal(this.element.textContent.trim(), 'template block text');
    });
  });
});
define("operator-docs/tests/integration/components/documentation/introduction-test", ["qunit", "ember-qunit", "@ember/test-helpers"], function (_qunit, _emberQunit, _testHelpers) {
  "use strict";

  (0, _qunit.module)('Integration | Component | documentation/introduction', function (hooks) {
    (0, _emberQunit.setupRenderingTest)(hooks);
    (0, _qunit.test)('it renders', async function (assert) {
      // Set any properties with this.set('myProperty', 'value');
      // Handle any actions with this.set('myAction', function(val) { ... });
      await (0, _testHelpers.render)(Ember.HTMLBars.template({
        "id": "GfR5vLyG",
        "block": "{\"symbols\":[],\"statements\":[[1,[23,\"documentation/introduction\"],false]],\"hasEval\":false}",
        "meta": {}
      }));
      assert.equal(this.element.textContent.trim(), ''); // Template block usage:

      await (0, _testHelpers.render)(Ember.HTMLBars.template({
        "id": "3LSEpSBr",
        "block": "{\"symbols\":[],\"statements\":[[0,\"\\n\"],[4,\"documentation/introduction\",null,null,{\"statements\":[[0,\"        template block text\\n\"]],\"parameters\":[]},null],[0,\"    \"]],\"hasEval\":false}",
        "meta": {}
      }));
      assert.equal(this.element.textContent.trim(), 'template block text');
    });
  });
});
define("operator-docs/tests/integration/components/documentation/on-premises-test", ["qunit", "ember-qunit", "@ember/test-helpers"], function (_qunit, _emberQunit, _testHelpers) {
  "use strict";

  (0, _qunit.module)('Integration | Component | documentation/on-premises', function (hooks) {
    (0, _emberQunit.setupRenderingTest)(hooks);
    (0, _qunit.test)('it renders', async function (assert) {
      // Set any properties with this.set('myProperty', 'value');
      // Handle any actions with this.set('myAction', function(val) { ... });
      await (0, _testHelpers.render)(Ember.HTMLBars.template({
        "id": "B4KEKNgi",
        "block": "{\"symbols\":[],\"statements\":[[1,[23,\"documentation/on-premises\"],false]],\"hasEval\":false}",
        "meta": {}
      }));
      assert.equal(this.element.textContent.trim(), ''); // Template block usage:

      await (0, _testHelpers.render)(Ember.HTMLBars.template({
        "id": "u3nxO/BJ",
        "block": "{\"symbols\":[],\"statements\":[[0,\"\\n\"],[4,\"documentation/on-premises\",null,null,{\"statements\":[[0,\"        template block text\\n\"]],\"parameters\":[]},null],[0,\"    \"]],\"hasEval\":false}",
        "meta": {}
      }));
      assert.equal(this.element.textContent.trim(), 'template block text');
    });
  });
});
define("operator-docs/tests/integration/components/documentation/overview-test", ["qunit", "ember-qunit", "@ember/test-helpers"], function (_qunit, _emberQunit, _testHelpers) {
  "use strict";

  (0, _qunit.module)('Integration | Component | documentation/overview', function (hooks) {
    (0, _emberQunit.setupRenderingTest)(hooks);
    (0, _qunit.test)('it renders', async function (assert) {
      // Set any properties with this.set('myProperty', 'value');
      // Handle any actions with this.set('myAction', function(val) { ... });
      await (0, _testHelpers.render)(Ember.HTMLBars.template({
        "id": "bFZT1Txa",
        "block": "{\"symbols\":[],\"statements\":[[1,[23,\"documentation/overview\"],false]],\"hasEval\":false}",
        "meta": {}
      }));
      assert.equal(this.element.textContent.trim(), ''); // Template block usage:

      await (0, _testHelpers.render)(Ember.HTMLBars.template({
        "id": "KE8Ajc05",
        "block": "{\"symbols\":[],\"statements\":[[0,\"\\n\"],[4,\"documentation/overview\",null,null,{\"statements\":[[0,\"        template block text\\n\"]],\"parameters\":[]},null],[0,\"    \"]],\"hasEval\":false}",
        "meta": {}
      }));
      assert.equal(this.element.textContent.trim(), 'template block text');
    });
  });
});
define("operator-docs/tests/integration/components/documentation/prerequisites-test", ["qunit", "ember-qunit", "@ember/test-helpers"], function (_qunit, _emberQunit, _testHelpers) {
  "use strict";

  (0, _qunit.module)('Integration | Component | documentation/prerequisites', function (hooks) {
    (0, _emberQunit.setupRenderingTest)(hooks);
    (0, _qunit.test)('it renders', async function (assert) {
      // Set any properties with this.set('myProperty', 'value');
      // Handle any actions with this.set('myAction', function(val) { ... });
      await (0, _testHelpers.render)(Ember.HTMLBars.template({
        "id": "4Z2zmYxB",
        "block": "{\"symbols\":[],\"statements\":[[1,[23,\"documentation/prerequisites\"],false]],\"hasEval\":false}",
        "meta": {}
      }));
      assert.equal(this.element.textContent.trim(), ''); // Template block usage:

      await (0, _testHelpers.render)(Ember.HTMLBars.template({
        "id": "ELVKURMw",
        "block": "{\"symbols\":[],\"statements\":[[0,\"\\n\"],[4,\"documentation/prerequisites\",null,null,{\"statements\":[[0,\"        template block text\\n\"]],\"parameters\":[]},null],[0,\"    \"]],\"hasEval\":false}",
        "meta": {}
      }));
      assert.equal(this.element.textContent.trim(), 'template block text');
    });
  });
});
define("operator-docs/tests/integration/components/polaris-form-test", ["qunit", "ember-qunit", "@ember/test-helpers"], function (_qunit, _emberQunit, _testHelpers) {
  "use strict";

  (0, _qunit.module)('Integration | Component | polaris-form', function (hooks) {
    (0, _emberQunit.setupRenderingTest)(hooks);
    (0, _qunit.test)('it renders', async function (assert) {
      // Set any properties with this.set('myProperty', 'value');
      // Handle any actions with this.set('myAction', function(val) { ... });
      await (0, _testHelpers.render)(Ember.HTMLBars.template({
        "id": "Y4a9ALGo",
        "block": "{\"symbols\":[],\"statements\":[[5,\"polaris-form\",[],[[],[]]]],\"hasEval\":false}",
        "meta": {}
      }));
      assert.equal(this.element.textContent.trim(), ''); // Template block usage:

      await (0, _testHelpers.render)(Ember.HTMLBars.template({
        "id": "sxb7ladN",
        "block": "{\"symbols\":[],\"statements\":[[0,\"\\n      \"],[5,\"polaris-form\",[],[[],[]],{\"statements\":[[0,\"\\n        template block text\\n      \"]],\"parameters\":[]}],[0,\"\\n    \"]],\"hasEval\":false}",
        "meta": {}
      }));
      assert.equal(this.element.textContent.trim(), 'template block text');
    });
  });
});
define("operator-docs/tests/integration/components/ui-brand-logo-test", ["qunit", "ember-qunit", "@ember/test-helpers"], function (_qunit, _emberQunit, _testHelpers) {
  "use strict";

  (0, _qunit.module)('Integration | Component | ui-brand-logo', function (hooks) {
    (0, _emberQunit.setupRenderingTest)(hooks);
    (0, _qunit.test)('it renders', async function (assert) {
      // Set any properties with this.set('myProperty', 'value');
      // Handle any actions with this.set('myAction', function(val) { ... });
      await (0, _testHelpers.render)(Ember.HTMLBars.template({
        "id": "wP49B6Zr",
        "block": "{\"symbols\":[],\"statements\":[[5,\"ui-brand-logo\",[],[[],[]]]],\"hasEval\":false}",
        "meta": {}
      }));
      assert.equal(this.element.textContent.trim(), ''); // Template block usage:

      await (0, _testHelpers.render)(Ember.HTMLBars.template({
        "id": "7FmWD7Jq",
        "block": "{\"symbols\":[],\"statements\":[[0,\"\\n      \"],[5,\"ui-brand-logo\",[],[[],[]],{\"statements\":[[0,\"\\n        template block text\\n      \"]],\"parameters\":[]}],[0,\"\\n    \"]],\"hasEval\":false}",
        "meta": {}
      }));
      assert.equal(this.element.textContent.trim(), 'template block text');
    });
  });
});
define("operator-docs/tests/integration/components/ui-head-test", ["qunit", "ember-qunit", "@ember/test-helpers"], function (_qunit, _emberQunit, _testHelpers) {
  "use strict";

  (0, _qunit.module)('Integration | Component | ui-head', function (hooks) {
    (0, _emberQunit.setupRenderingTest)(hooks);
    (0, _qunit.test)('it renders', async function (assert) {
      // Set any properties with this.set('myProperty', 'value');
      // Handle any actions with this.set('myAction', function(val) { ... });
      await (0, _testHelpers.render)(Ember.HTMLBars.template({
        "id": "SWNDu4U5",
        "block": "{\"symbols\":[],\"statements\":[[5,\"ui-head\",[],[[],[]]]],\"hasEval\":false}",
        "meta": {}
      }));
      assert.equal(this.element.textContent.trim(), ''); // Template block usage:

      await (0, _testHelpers.render)(Ember.HTMLBars.template({
        "id": "YnGRbyh1",
        "block": "{\"symbols\":[],\"statements\":[[0,\"\\n      \"],[5,\"ui-head\",[],[[],[]],{\"statements\":[[0,\"\\n        template block text\\n      \"]],\"parameters\":[]}],[0,\"\\n    \"]],\"hasEval\":false}",
        "meta": {}
      }));
      assert.equal(this.element.textContent.trim(), 'template block text');
    });
  });
});
define("operator-docs/tests/integration/components/ui-mobile-header-test", ["qunit", "ember-qunit", "@ember/test-helpers"], function (_qunit, _emberQunit, _testHelpers) {
  "use strict";

  (0, _qunit.module)('Integration | Component | ui-mobile-header', function (hooks) {
    (0, _emberQunit.setupRenderingTest)(hooks);
    (0, _qunit.test)('it renders', async function (assert) {
      // Set any properties with this.set('myProperty', 'value');
      // Handle any actions with this.set('myAction', function(val) { ... });
      await (0, _testHelpers.render)(Ember.HTMLBars.template({
        "id": "Pd3mj28I",
        "block": "{\"symbols\":[],\"statements\":[[5,\"ui-mobile-header\",[],[[],[]]]],\"hasEval\":false}",
        "meta": {}
      }));
      assert.equal(this.element.textContent.trim(), ''); // Template block usage:

      await (0, _testHelpers.render)(Ember.HTMLBars.template({
        "id": "w1Zic3+g",
        "block": "{\"symbols\":[],\"statements\":[[0,\"\\n      \"],[5,\"ui-mobile-header\",[],[[],[]],{\"statements\":[[0,\"\\n        template block text\\n      \"]],\"parameters\":[]}],[0,\"\\n    \"]],\"hasEval\":false}",
        "meta": {}
      }));
      assert.equal(this.element.textContent.trim(), 'template block text');
    });
  });
});
define("operator-docs/tests/integration/components/ui-nav-bar-test", ["qunit", "ember-qunit", "@ember/test-helpers"], function (_qunit, _emberQunit, _testHelpers) {
  "use strict";

  (0, _qunit.module)('Integration | Component | ui-nav-bar', function (hooks) {
    (0, _emberQunit.setupRenderingTest)(hooks);
    (0, _qunit.test)('it renders', async function (assert) {
      // Set any properties with this.set('myProperty', 'value');
      // Handle any actions with this.set('myAction', function(val) { ... });
      await (0, _testHelpers.render)(Ember.HTMLBars.template({
        "id": "DtwaNbVv",
        "block": "{\"symbols\":[],\"statements\":[[5,\"ui-nav-bar\",[],[[],[]]]],\"hasEval\":false}",
        "meta": {}
      }));
      assert.equal(this.element.textContent.trim(), ''); // Template block usage:

      await (0, _testHelpers.render)(Ember.HTMLBars.template({
        "id": "2dKpv5r3",
        "block": "{\"symbols\":[],\"statements\":[[0,\"\\n      \"],[5,\"ui-nav-bar\",[],[[],[]],{\"statements\":[[0,\"\\n        template block text\\n      \"]],\"parameters\":[]}],[0,\"\\n    \"]],\"hasEval\":false}",
        "meta": {}
      }));
      assert.equal(this.element.textContent.trim(), 'template block text');
    });
  });
});
define("operator-docs/tests/integration/components/ui/help-test", ["qunit", "ember-qunit", "@ember/test-helpers"], function (_qunit, _emberQunit, _testHelpers) {
  "use strict";

  (0, _qunit.module)('Integration | Component | ui/help', function (hooks) {
    (0, _emberQunit.setupRenderingTest)(hooks);
    (0, _qunit.test)('it renders', async function (assert) {
      // Set any properties with this.set('myProperty', 'value');
      // Handle any actions with this.set('myAction', function(val) { ... });
      await (0, _testHelpers.render)(Ember.HTMLBars.template({
        "id": "97a8CsQS",
        "block": "{\"symbols\":[],\"statements\":[[1,[23,\"ui/help\"],false]],\"hasEval\":false}",
        "meta": {}
      }));
      assert.equal(this.element.textContent.trim(), ''); // Template block usage:

      await (0, _testHelpers.render)(Ember.HTMLBars.template({
        "id": "iyP0/PWr",
        "block": "{\"symbols\":[],\"statements\":[[0,\"\\n\"],[4,\"ui/help\",null,null,{\"statements\":[[0,\"        template block text\\n\"]],\"parameters\":[]},null],[0,\"    \"]],\"hasEval\":false}",
        "meta": {}
      }));
      assert.equal(this.element.textContent.trim(), 'template block text');
    });
  });
});
define("operator-docs/tests/integration/components/ui/home-test", ["qunit", "ember-qunit", "@ember/test-helpers"], function (_qunit, _emberQunit, _testHelpers) {
  "use strict";

  (0, _qunit.module)('Integration | Component | ui/home', function (hooks) {
    (0, _emberQunit.setupRenderingTest)(hooks);
    (0, _qunit.test)('it renders', async function (assert) {
      // Set any properties with this.set('myProperty', 'value');
      // Handle any actions with this.set('myAction', function(val) { ... });
      await (0, _testHelpers.render)(Ember.HTMLBars.template({
        "id": "nRgQ90pw",
        "block": "{\"symbols\":[],\"statements\":[[1,[23,\"ui/home\"],false]],\"hasEval\":false}",
        "meta": {}
      }));
      assert.equal(this.element.textContent.trim(), ''); // Template block usage:

      await (0, _testHelpers.render)(Ember.HTMLBars.template({
        "id": "KezOB9HG",
        "block": "{\"symbols\":[],\"statements\":[[0,\"\\n\"],[4,\"ui/home\",null,null,{\"statements\":[[0,\"        template block text\\n\"]],\"parameters\":[]},null],[0,\"    \"]],\"hasEval\":false}",
        "meta": {}
      }));
      assert.equal(this.element.textContent.trim(), 'template block text');
    });
  });
});
define("operator-docs/tests/integration/components/ui/operator-test", ["qunit", "ember-qunit", "@ember/test-helpers"], function (_qunit, _emberQunit, _testHelpers) {
  "use strict";

  (0, _qunit.module)('Integration | Component | ui/operator', function (hooks) {
    (0, _emberQunit.setupRenderingTest)(hooks);
    (0, _qunit.test)('it renders', async function (assert) {
      // Set any properties with this.set('myProperty', 'value');
      // Handle any actions with this.set('myAction', function(val) { ... });
      await (0, _testHelpers.render)(Ember.HTMLBars.template({
        "id": "D3KNFfbD",
        "block": "{\"symbols\":[],\"statements\":[[1,[23,\"ui/operator\"],false]],\"hasEval\":false}",
        "meta": {}
      }));
      assert.equal(this.element.textContent.trim(), ''); // Template block usage:

      await (0, _testHelpers.render)(Ember.HTMLBars.template({
        "id": "ECBko71t",
        "block": "{\"symbols\":[],\"statements\":[[0,\"\\n\"],[4,\"ui/operator\",null,null,{\"statements\":[[0,\"        template block text\\n\"]],\"parameters\":[]},null],[0,\"    \"]],\"hasEval\":false}",
        "meta": {}
      }));
      assert.equal(this.element.textContent.trim(), 'template block text');
    });
  });
});
define("operator-docs/tests/lint/app.lint-test", [], function () {
  "use strict";

  QUnit.module('ESLint | app');
  QUnit.test('app.js', function (assert) {
    assert.expect(1);
    assert.ok(true, 'app.js should pass ESLint\n\n');
  });
  QUnit.test('components/black-duck-form.js', function (assert) {
    assert.expect(1);
    assert.ok(false, 'components/black-duck-form.js should pass ESLint\n\n4:5 - Only string, number, symbol, boolean, null, undefined, and function are allowed as default properties (ember/avoid-leaking-state-in-ember-objects)\n31:13 - \'$\' is not defined. (no-undef)');
  });
  QUnit.test('components/documentation-navbar.js', function (assert) {
    assert.expect(1);
    assert.ok(true, 'components/documentation-navbar.js should pass ESLint\n\n');
  });
  QUnit.test('components/documentation/aks.js', function (assert) {
    assert.expect(1);
    assert.ok(true, 'components/documentation/aks.js should pass ESLint\n\n');
  });
  QUnit.test('components/documentation/deploy-operator.js', function (assert) {
    assert.expect(1);
    assert.ok(true, 'components/documentation/deploy-operator.js should pass ESLint\n\n');
  });
  QUnit.test('components/documentation/eks.js', function (assert) {
    assert.expect(1);
    assert.ok(true, 'components/documentation/eks.js should pass ESLint\n\n');
  });
  QUnit.test('components/documentation/gke.js', function (assert) {
    assert.expect(1);
    assert.ok(true, 'components/documentation/gke.js should pass ESLint\n\n');
  });
  QUnit.test('components/documentation/home.js', function (assert) {
    assert.expect(1);
    assert.ok(true, 'components/documentation/home.js should pass ESLint\n\n');
  });
  QUnit.test('components/documentation/introduction.js', function (assert) {
    assert.expect(1);
    assert.ok(true, 'components/documentation/introduction.js should pass ESLint\n\n');
  });
  QUnit.test('components/documentation/on-premises.js', function (assert) {
    assert.expect(1);
    assert.ok(true, 'components/documentation/on-premises.js should pass ESLint\n\n');
  });
  QUnit.test('components/documentation/overview.js', function (assert) {
    assert.expect(1);
    assert.ok(true, 'components/documentation/overview.js should pass ESLint\n\n');
  });
  QUnit.test('components/documentation/prerequisites.js', function (assert) {
    assert.expect(1);
    assert.ok(true, 'components/documentation/prerequisites.js should pass ESLint\n\n');
  });
  QUnit.test('components/polaris-form.js', function (assert) {
    assert.expect(1);
    assert.ok(false, 'components/polaris-form.js should pass ESLint\n\n4:5 - Only string, number, symbol, boolean, null, undefined, and function are allowed as default properties (ember/avoid-leaking-state-in-ember-objects)\n19:13 - \'$\' is not defined. (no-undef)');
  });
  QUnit.test('components/ui-brand-logo.js', function (assert) {
    assert.expect(1);
    assert.ok(true, 'components/ui-brand-logo.js should pass ESLint\n\n');
  });
  QUnit.test('components/ui-head.js', function (assert) {
    assert.expect(1);
    assert.ok(true, 'components/ui-head.js should pass ESLint\n\n');
  });
  QUnit.test('components/ui-mobile-header.js', function (assert) {
    assert.expect(1);
    assert.ok(true, 'components/ui-mobile-header.js should pass ESLint\n\n');
  });
  QUnit.test('components/ui-nav-bar.js', function (assert) {
    assert.expect(1);
    assert.ok(true, 'components/ui-nav-bar.js should pass ESLint\n\n');
  });
  QUnit.test('components/ui/help.js', function (assert) {
    assert.expect(1);
    assert.ok(true, 'components/ui/help.js should pass ESLint\n\n');
  });
  QUnit.test('components/ui/home.js', function (assert) {
    assert.expect(1);
    assert.ok(true, 'components/ui/home.js should pass ESLint\n\n');
  });
  QUnit.test('components/ui/operator.js', function (assert) {
    assert.expect(1);
    assert.ok(true, 'components/ui/operator.js should pass ESLint\n\n');
  });
  QUnit.test('resolver.js', function (assert) {
    assert.expect(1);
    assert.ok(true, 'resolver.js should pass ESLint\n\n');
  });
  QUnit.test('router.js', function (assert) {
    assert.expect(1);
    assert.ok(true, 'router.js should pass ESLint\n\n');
  });
  QUnit.test('routes/documentation.js', function (assert) {
    assert.expect(1);
    assert.ok(true, 'routes/documentation.js should pass ESLint\n\n');
  });
  QUnit.test('routes/documentation/aks.js', function (assert) {
    assert.expect(1);
    assert.ok(true, 'routes/documentation/aks.js should pass ESLint\n\n');
  });
  QUnit.test('routes/documentation/deploy-operator.js', function (assert) {
    assert.expect(1);
    assert.ok(true, 'routes/documentation/deploy-operator.js should pass ESLint\n\n');
  });
  QUnit.test('routes/documentation/eks.js', function (assert) {
    assert.expect(1);
    assert.ok(true, 'routes/documentation/eks.js should pass ESLint\n\n');
  });
  QUnit.test('routes/documentation/gke.js', function (assert) {
    assert.expect(1);
    assert.ok(true, 'routes/documentation/gke.js should pass ESLint\n\n');
  });
  QUnit.test('routes/documentation/home.js', function (assert) {
    assert.expect(1);
    assert.ok(true, 'routes/documentation/home.js should pass ESLint\n\n');
  });
  QUnit.test('routes/documentation/on-premises.js', function (assert) {
    assert.expect(1);
    assert.ok(true, 'routes/documentation/on-premises.js should pass ESLint\n\n');
  });
  QUnit.test('routes/documentation/overview.js', function (assert) {
    assert.expect(1);
    assert.ok(true, 'routes/documentation/overview.js should pass ESLint\n\n');
  });
  QUnit.test('routes/documentation/prerequisites.js', function (assert) {
    assert.expect(1);
    assert.ok(true, 'routes/documentation/prerequisites.js should pass ESLint\n\n');
  });
  QUnit.test('routes/ui.js', function (assert) {
    assert.expect(1);
    assert.ok(true, 'routes/ui.js should pass ESLint\n\n');
  });
  QUnit.test('routes/ui/deploy-black-duck.js', function (assert) {
    assert.expect(1);
    assert.ok(true, 'routes/ui/deploy-black-duck.js should pass ESLint\n\n');
  });
  QUnit.test('routes/ui/deploy-polaris.js', function (assert) {
    assert.expect(1);
    assert.ok(true, 'routes/ui/deploy-polaris.js should pass ESLint\n\n');
  });
  QUnit.test('routes/ui/help.js', function (assert) {
    assert.expect(1);
    assert.ok(true, 'routes/ui/help.js should pass ESLint\n\n');
  });
  QUnit.test('routes/ui/home.js', function (assert) {
    assert.expect(1);
    assert.ok(true, 'routes/ui/home.js should pass ESLint\n\n');
  });
  QUnit.test('routes/ui/operator.js', function (assert) {
    assert.expect(1);
    assert.ok(true, 'routes/ui/operator.js should pass ESLint\n\n');
  });
});
define("operator-docs/tests/lint/templates.template.lint-test", [], function () {
  "use strict";

  QUnit.module('TemplateLint');
  QUnit.test('operator-docs/templates/application.hbs', function (assert) {
    assert.expect(1);
    assert.ok(true, 'operator-docs/templates/application.hbs should pass TemplateLint.\n\n');
  });
  QUnit.test('operator-docs/templates/components/black-duck-form.hbs', function (assert) {
    assert.expect(1);
    assert.ok(false, 'operator-docs/templates/components/black-duck-form.hbs should pass TemplateLint.\n\noperator-docs/templates/components/black-duck-form.hbs\n  3:4  error  Incorrect indentation for `<form>` beginning at L3:C4. Expected `<form>` to be at an indentation of 2 but was found at 4.  block-indentation\n  4:8  error  Incorrect indentation for `<div>` beginning at L4:C8. Expected `<div>` to be at an indentation of 6 but was found at 8.  block-indentation\n  10:8  error  Incorrect indentation for `<div>` beginning at L10:C8. Expected `<div>` to be at an indentation of 6 but was found at 8.  block-indentation\n  16:8  error  Incorrect indentation for `<div>` beginning at L16:C8. Expected `<div>` to be at an indentation of 6 but was found at 8.  block-indentation\n  22:8  error  Incorrect indentation for `<div>` beginning at L22:C8. Expected `<div>` to be at an indentation of 6 but was found at 8.  block-indentation\n  28:8  error  Incorrect indentation for `<div>` beginning at L28:C8. Expected `<div>` to be at an indentation of 6 but was found at 8.  block-indentation\n  34:8  error  Incorrect indentation for `<div>` beginning at L34:C8. Expected `<div>` to be at an indentation of 6 but was found at 8.  block-indentation\n  40:8  error  Incorrect indentation for `<div>` beginning at L40:C8. Expected `<div>` to be at an indentation of 6 but was found at 8.  block-indentation\n  46:8  error  Incorrect indentation for `<div>` beginning at L46:C8. Expected `<div>` to be at an indentation of 6 but was found at 8.  block-indentation\n  53:8  error  Incorrect indentation for `<div>` beginning at L53:C8. Expected `<div>` to be at an indentation of 6 but was found at 8.  block-indentation\n  59:8  error  Incorrect indentation for `<div>` beginning at L59:C8. Expected `<div>` to be at an indentation of 6 but was found at 8.  block-indentation\n  65:8  error  Incorrect indentation for `<div>` beginning at L65:C8. Expected `<div>` to be at an indentation of 6 but was found at 8.  block-indentation\n  71:8  error  Incorrect indentation for `<div>` beginning at L71:C8. Expected `<div>` to be at an indentation of 6 but was found at 8.  block-indentation\n  77:8  error  Incorrect indentation for `<div>` beginning at L77:C8. Expected `<div>` to be at an indentation of 6 but was found at 8.  block-indentation\n  83:8  error  Incorrect indentation for `<div>` beginning at L83:C8. Expected `<div>` to be at an indentation of 6 but was found at 8.  block-indentation\n  89:8  error  Incorrect indentation for `<div>` beginning at L89:C8. Expected `<div>` to be at an indentation of 6 but was found at 8.  block-indentation\n  95:8  error  Incorrect indentation for `<div>` beginning at L95:C8. Expected `<div>` to be at an indentation of 6 but was found at 8.  block-indentation\n  101:8  error  Incorrect indentation for `<div>` beginning at L101:C8. Expected `<div>` to be at an indentation of 6 but was found at 8.  block-indentation\n  107:8  error  Incorrect indentation for `<div>` beginning at L107:C8. Expected `<div>` to be at an indentation of 6 but was found at 8.  block-indentation\n  113:8  error  Incorrect indentation for `<div>` beginning at L113:C8. Expected `<div>` to be at an indentation of 6 but was found at 8.  block-indentation\n  119:8  error  Incorrect indentation for `<div>` beginning at L119:C8. Expected `<div>` to be at an indentation of 6 but was found at 8.  block-indentation\n  125:8  error  Incorrect indentation for `<div>` beginning at L125:C8. Expected `<div>` to be at an indentation of 6 but was found at 8.  block-indentation\n  131:8  error  Incorrect indentation for `<div>` beginning at L131:C8. Expected `<div>` to be at an indentation of 6 but was found at 8.  block-indentation\n  141:8  error  Incorrect indentation for `<div>` beginning at L141:C8. Expected `<div>` to be at an indentation of 6 but was found at 8.  block-indentation\n  5:12  error  Incorrect indentation for `<label>` beginning at L5:C12. Expected `<label>` to be at an indentation of 10 but was found at 12.  block-indentation\n  6:12  error  Incorrect indentation for `<div>` beginning at L6:C12. Expected `<div>` to be at an indentation of 10 but was found at 12.  block-indentation\n  7:16  error  Incorrect indentation for `{{input}}` beginning at L7:C16. Expected `{{input}}` to be at an indentation of 14 but was found at 16.  block-indentation\n  11:12  error  Incorrect indentation for `<label>` beginning at L11:C12. Expected `<label>` to be at an indentation of 10 but was found at 12.  block-indentation\n  12:12  error  Incorrect indentation for `<div>` beginning at L12:C12. Expected `<div>` to be at an indentation of 10 but was found at 12.  block-indentation\n  13:16  error  Incorrect indentation for `{{input}}` beginning at L13:C16. Expected `{{input}}` to be at an indentation of 14 but was found at 16.  block-indentation\n  17:12  error  Incorrect indentation for `<label>` beginning at L17:C12. Expected `<label>` to be at an indentation of 10 but was found at 12.  block-indentation\n  18:12  error  Incorrect indentation for `<div>` beginning at L18:C12. Expected `<div>` to be at an indentation of 10 but was found at 12.  block-indentation\n  19:16  error  Incorrect indentation for `{{input}}` beginning at L19:C16. Expected `{{input}}` to be at an indentation of 14 but was found at 16.  block-indentation\n  23:12  error  Incorrect indentation for `<label>` beginning at L23:C12. Expected `<label>` to be at an indentation of 10 but was found at 12.  block-indentation\n  24:12  error  Incorrect indentation for `<div>` beginning at L24:C12. Expected `<div>` to be at an indentation of 10 but was found at 12.  block-indentation\n  25:16  error  Incorrect indentation for `{{input}}` beginning at L25:C16. Expected `{{input}}` to be at an indentation of 14 but was found at 16.  block-indentation\n  29:12  error  Incorrect indentation for `<label>` beginning at L29:C12. Expected `<label>` to be at an indentation of 10 but was found at 12.  block-indentation\n  30:12  error  Incorrect indentation for `<div>` beginning at L30:C12. Expected `<div>` to be at an indentation of 10 but was found at 12.  block-indentation\n  31:16  error  Incorrect indentation for `<Input>` beginning at L31:C16. Expected `<Input>` to be at an indentation of 14 but was found at 16.  block-indentation\n  35:12  error  Incorrect indentation for `<label>` beginning at L35:C12. Expected `<label>` to be at an indentation of 10 but was found at 12.  block-indentation\n  36:12  error  Incorrect indentation for `<div>` beginning at L36:C12. Expected `<div>` to be at an indentation of 10 but was found at 12.  block-indentation\n  37:16  error  Incorrect indentation for `{{input}}` beginning at L37:C16. Expected `{{input}}` to be at an indentation of 14 but was found at 16.  block-indentation\n  41:12  error  Incorrect indentation for `<label>` beginning at L41:C12. Expected `<label>` to be at an indentation of 10 but was found at 12.  block-indentation\n  42:12  error  Incorrect indentation for `<div>` beginning at L42:C12. Expected `<div>` to be at an indentation of 10 but was found at 12.  block-indentation\n  43:16  error  Incorrect indentation for `{{input}}` beginning at L43:C16. Expected `{{input}}` to be at an indentation of 14 but was found at 16.  block-indentation\n  47:12  error  Incorrect indentation for `<label>` beginning at L47:C12. Expected `<label>` to be at an indentation of 10 but was found at 12.  block-indentation\n  49:12  error  Incorrect indentation for `<div>` beginning at L49:C12. Expected `<div>` to be at an indentation of 10 but was found at 12.  block-indentation\n  48:44  error  Incorrect indentation for `label` beginning at L47:C12. Expected `</label>` ending at L48:C44 to be at an indentation of 12 but was found at 36.  block-indentation\n  47:72  error  Incorrect indentation for `Black Duck Type (OpsSight\n                blackDuckConfigific)` beginning at L47:C72. Expected `Black Duck Type (OpsSight\n                blackDuckConfigific)` to be at an indentation of 14 but was found at 72.  block-indentation\n  50:16  error  Incorrect indentation for `{{input}}` beginning at L50:C16. Expected `{{input}}` to be at an indentation of 14 but was found at 16.  block-indentation\n  54:12  error  Incorrect indentation for `<label>` beginning at L54:C12. Expected `<label>` to be at an indentation of 10 but was found at 12.  block-indentation\n  55:12  error  Incorrect indentation for `<div>` beginning at L55:C12. Expected `<div>` to be at an indentation of 10 but was found at 12.  block-indentation\n  56:16  error  Incorrect indentation for `<Input>` beginning at L56:C16. Expected `<Input>` to be at an indentation of 14 but was found at 16.  block-indentation\n  60:12  error  Incorrect indentation for `<label>` beginning at L60:C12. Expected `<label>` to be at an indentation of 10 but was found at 12.  block-indentation\n  61:12  error  Incorrect indentation for `<div>` beginning at L61:C12. Expected `<div>` to be at an indentation of 10 but was found at 12.  block-indentation\n  62:16  error  Incorrect indentation for `<Input>` beginning at L62:C16. Expected `<Input>` to be at an indentation of 14 but was found at 16.  block-indentation\n  66:12  error  Incorrect indentation for `<label>` beginning at L66:C12. Expected `<label>` to be at an indentation of 10 but was found at 12.  block-indentation\n  67:12  error  Incorrect indentation for `<div>` beginning at L67:C12. Expected `<div>` to be at an indentation of 10 but was found at 12.  block-indentation\n  68:16  error  Incorrect indentation for `<Input>` beginning at L68:C16. Expected `<Input>` to be at an indentation of 14 but was found at 16.  block-indentation\n  72:12  error  Incorrect indentation for `<label>` beginning at L72:C12. Expected `<label>` to be at an indentation of 10 but was found at 12.  block-indentation\n  73:12  error  Incorrect indentation for `<div>` beginning at L73:C12. Expected `<div>` to be at an indentation of 10 but was found at 12.  block-indentation\n  74:16  error  Incorrect indentation for `<Input>` beginning at L74:C16. Expected `<Input>` to be at an indentation of 14 but was found at 16.  block-indentation\n  78:12  error  Incorrect indentation for `<label>` beginning at L78:C12. Expected `<label>` to be at an indentation of 10 but was found at 12.  block-indentation\n  79:12  error  Incorrect indentation for `<div>` beginning at L79:C12. Expected `<div>` to be at an indentation of 10 but was found at 12.  block-indentation\n  80:16  error  Incorrect indentation for `{{input}}` beginning at L80:C16. Expected `{{input}}` to be at an indentation of 14 but was found at 16.  block-indentation\n  84:12  error  Incorrect indentation for `<label>` beginning at L84:C12. Expected `<label>` to be at an indentation of 10 but was found at 12.  block-indentation\n  85:12  error  Incorrect indentation for `<div>` beginning at L85:C12. Expected `<div>` to be at an indentation of 10 but was found at 12.  block-indentation\n  86:16  error  Incorrect indentation for `{{input}}` beginning at L86:C16. Expected `{{input}}` to be at an indentation of 14 but was found at 16.  block-indentation\n  90:12  error  Incorrect indentation for `<label>` beginning at L90:C12. Expected `<label>` to be at an indentation of 10 but was found at 12.  block-indentation\n  91:12  error  Incorrect indentation for `<div>` beginning at L91:C12. Expected `<div>` to be at an indentation of 10 but was found at 12.  block-indentation\n  92:16  error  Incorrect indentation for `{{input}}` beginning at L92:C16. Expected `{{input}}` to be at an indentation of 14 but was found at 16.  block-indentation\n  96:12  error  Incorrect indentation for `<label>` beginning at L96:C12. Expected `<label>` to be at an indentation of 10 but was found at 12.  block-indentation\n  97:12  error  Incorrect indentation for `<div>` beginning at L97:C12. Expected `<div>` to be at an indentation of 10 but was found at 12.  block-indentation\n  98:16  error  Incorrect indentation for `<Input>` beginning at L98:C16. Expected `<Input>` to be at an indentation of 14 but was found at 16.  block-indentation\n  102:12  error  Incorrect indentation for `<label>` beginning at L102:C12. Expected `<label>` to be at an indentation of 10 but was found at 12.  block-indentation\n  103:12  error  Incorrect indentation for `<div>` beginning at L103:C12. Expected `<div>` to be at an indentation of 10 but was found at 12.  block-indentation\n  104:16  error  Incorrect indentation for `{{input}}` beginning at L104:C16. Expected `{{input}}` to be at an indentation of 14 but was found at 16.  block-indentation\n  108:12  error  Incorrect indentation for `<label>` beginning at L108:C12. Expected `<label>` to be at an indentation of 10 but was found at 12.  block-indentation\n  109:12  error  Incorrect indentation for `<div>` beginning at L109:C12. Expected `<div>` to be at an indentation of 10 but was found at 12.  block-indentation\n  110:16  error  Incorrect indentation for `{{input}}` beginning at L110:C16. Expected `{{input}}` to be at an indentation of 14 but was found at 16.  block-indentation\n  114:12  error  Incorrect indentation for `<label>` beginning at L114:C12. Expected `<label>` to be at an indentation of 10 but was found at 12.  block-indentation\n  115:12  error  Incorrect indentation for `<div>` beginning at L115:C12. Expected `<div>` to be at an indentation of 10 but was found at 12.  block-indentation\n  116:16  error  Incorrect indentation for `{{input}}` beginning at L116:C16. Expected `{{input}}` to be at an indentation of 14 but was found at 16.  block-indentation\n  120:12  error  Incorrect indentation for `<label>` beginning at L120:C12. Expected `<label>` to be at an indentation of 10 but was found at 12.  block-indentation\n  121:12  error  Incorrect indentation for `<div>` beginning at L121:C12. Expected `<div>` to be at an indentation of 10 but was found at 12.  block-indentation\n  122:16  error  Incorrect indentation for `{{input}}` beginning at L122:C16. Expected `{{input}}` to be at an indentation of 14 but was found at 16.  block-indentation\n  126:12  error  Incorrect indentation for `<label>` beginning at L126:C12. Expected `<label>` to be at an indentation of 10 but was found at 12.  block-indentation\n  127:12  error  Incorrect indentation for `<div>` beginning at L127:C12. Expected `<div>` to be at an indentation of 10 but was found at 12.  block-indentation\n  128:16  error  Incorrect indentation for `<Input>` beginning at L128:C16. Expected `<Input>` to be at an indentation of 14 but was found at 16.  block-indentation\n  132:12  error  Incorrect indentation for `<label>` beginning at L132:C12. Expected `<label>` to be at an indentation of 10 but was found at 12.  block-indentation\n  133:12  error  Incorrect indentation for `<div>` beginning at L133:C12. Expected `<div>` to be at an indentation of 10 but was found at 12.  block-indentation\n  135:22  error  Incorrect indentation for `div` beginning at L133:C12. Expected `</div>` ending at L135:C22 to be at an indentation of 12 but was found at 16.  block-indentation\n  134:16  error  Incorrect indentation for `<Textarea>` beginning at L134:C16. Expected `<Textarea>` to be at an indentation of 14 but was found at 16.  block-indentation\n  142:12  error  Incorrect indentation for `<div>` beginning at L142:C12. Expected `<div>` to be at an indentation of 10 but was found at 12.  block-indentation\n  143:16  error  Incorrect indentation for `<button>` beginning at L143:C16. Expected `<button>` to be at an indentation of 14 but was found at 16.  block-indentation\n');
  });
  QUnit.test('operator-docs/templates/components/documentation-navbar.hbs', function (assert) {
    assert.expect(1);
    assert.ok(true, 'operator-docs/templates/components/documentation-navbar.hbs should pass TemplateLint.\n\n');
  });
  QUnit.test('operator-docs/templates/components/documentation/aks.hbs', function (assert) {
    assert.expect(1);
    assert.ok(true, 'operator-docs/templates/components/documentation/aks.hbs should pass TemplateLint.\n\n');
  });
  QUnit.test('operator-docs/templates/components/documentation/deploy-operator.hbs', function (assert) {
    assert.expect(1);
    assert.ok(true, 'operator-docs/templates/components/documentation/deploy-operator.hbs should pass TemplateLint.\n\n');
  });
  QUnit.test('operator-docs/templates/components/documentation/eks.hbs', function (assert) {
    assert.expect(1);
    assert.ok(true, 'operator-docs/templates/components/documentation/eks.hbs should pass TemplateLint.\n\n');
  });
  QUnit.test('operator-docs/templates/components/documentation/gke.hbs', function (assert) {
    assert.expect(1);
    assert.ok(true, 'operator-docs/templates/components/documentation/gke.hbs should pass TemplateLint.\n\n');
  });
  QUnit.test('operator-docs/templates/components/documentation/home.hbs', function (assert) {
    assert.expect(1);
    assert.ok(true, 'operator-docs/templates/components/documentation/home.hbs should pass TemplateLint.\n\n');
  });
  QUnit.test('operator-docs/templates/components/documentation/introduction.hbs', function (assert) {
    assert.expect(1);
    assert.ok(true, 'operator-docs/templates/components/documentation/introduction.hbs should pass TemplateLint.\n\n');
  });
  QUnit.test('operator-docs/templates/components/documentation/on-premises.hbs', function (assert) {
    assert.expect(1);
    assert.ok(true, 'operator-docs/templates/components/documentation/on-premises.hbs should pass TemplateLint.\n\n');
  });
  QUnit.test('operator-docs/templates/components/documentation/overview.hbs', function (assert) {
    assert.expect(1);
    assert.ok(true, 'operator-docs/templates/components/documentation/overview.hbs should pass TemplateLint.\n\n');
  });
  QUnit.test('operator-docs/templates/components/documentation/prerequisites.hbs', function (assert) {
    assert.expect(1);
    assert.ok(true, 'operator-docs/templates/components/documentation/prerequisites.hbs should pass TemplateLint.\n\n');
  });
  QUnit.test('operator-docs/templates/components/polaris-form.hbs', function (assert) {
    assert.expect(1);
    assert.ok(false, 'operator-docs/templates/components/polaris-form.hbs should pass TemplateLint.\n\noperator-docs/templates/components/polaris-form.hbs\n  89:32  error  Closing tag `span` (on line 92) did not match last open tag `Input` (on line 89).  undefined\n');
  });
  QUnit.test('operator-docs/templates/components/ui-brand-logo.hbs', function (assert) {
    assert.expect(1);
    assert.ok(true, 'operator-docs/templates/components/ui-brand-logo.hbs should pass TemplateLint.\n\n');
  });
  QUnit.test('operator-docs/templates/components/ui-head.hbs', function (assert) {
    assert.expect(1);
    assert.ok(true, 'operator-docs/templates/components/ui-head.hbs should pass TemplateLint.\n\n');
  });
  QUnit.test('operator-docs/templates/components/ui-mobile-header.hbs', function (assert) {
    assert.expect(1);
    assert.ok(true, 'operator-docs/templates/components/ui-mobile-header.hbs should pass TemplateLint.\n\n');
  });
  QUnit.test('operator-docs/templates/components/ui-nav-bar.hbs', function (assert) {
    assert.expect(1);
    assert.ok(true, 'operator-docs/templates/components/ui-nav-bar.hbs should pass TemplateLint.\n\n');
  });
  QUnit.test('operator-docs/templates/components/ui/help.hbs', function (assert) {
    assert.expect(1);
    assert.ok(true, 'operator-docs/templates/components/ui/help.hbs should pass TemplateLint.\n\n');
  });
  QUnit.test('operator-docs/templates/components/ui/home.hbs', function (assert) {
    assert.expect(1);
    assert.ok(true, 'operator-docs/templates/components/ui/home.hbs should pass TemplateLint.\n\n');
  });
  QUnit.test('operator-docs/templates/components/ui/operator.hbs', function (assert) {
    assert.expect(1);
    assert.ok(true, 'operator-docs/templates/components/ui/operator.hbs should pass TemplateLint.\n\n');
  });
  QUnit.test('operator-docs/templates/documentation.hbs', function (assert) {
    assert.expect(1);
    assert.ok(true, 'operator-docs/templates/documentation.hbs should pass TemplateLint.\n\n');
  });
  QUnit.test('operator-docs/templates/documentation/aks.hbs', function (assert) {
    assert.expect(1);
    assert.ok(true, 'operator-docs/templates/documentation/aks.hbs should pass TemplateLint.\n\n');
  });
  QUnit.test('operator-docs/templates/documentation/deploy-operator.hbs', function (assert) {
    assert.expect(1);
    assert.ok(true, 'operator-docs/templates/documentation/deploy-operator.hbs should pass TemplateLint.\n\n');
  });
  QUnit.test('operator-docs/templates/documentation/eks.hbs', function (assert) {
    assert.expect(1);
    assert.ok(true, 'operator-docs/templates/documentation/eks.hbs should pass TemplateLint.\n\n');
  });
  QUnit.test('operator-docs/templates/documentation/gke.hbs', function (assert) {
    assert.expect(1);
    assert.ok(true, 'operator-docs/templates/documentation/gke.hbs should pass TemplateLint.\n\n');
  });
  QUnit.test('operator-docs/templates/documentation/home.hbs', function (assert) {
    assert.expect(1);
    assert.ok(true, 'operator-docs/templates/documentation/home.hbs should pass TemplateLint.\n\n');
  });
  QUnit.test('operator-docs/templates/documentation/on-premises.hbs', function (assert) {
    assert.expect(1);
    assert.ok(true, 'operator-docs/templates/documentation/on-premises.hbs should pass TemplateLint.\n\n');
  });
  QUnit.test('operator-docs/templates/documentation/overview.hbs', function (assert) {
    assert.expect(1);
    assert.ok(true, 'operator-docs/templates/documentation/overview.hbs should pass TemplateLint.\n\n');
  });
  QUnit.test('operator-docs/templates/documentation/prerequisites.hbs', function (assert) {
    assert.expect(1);
    assert.ok(true, 'operator-docs/templates/documentation/prerequisites.hbs should pass TemplateLint.\n\n');
  });
  QUnit.test('operator-docs/templates/ui.hbs', function (assert) {
    assert.expect(1);
    assert.ok(true, 'operator-docs/templates/ui.hbs should pass TemplateLint.\n\n');
  });
  QUnit.test('operator-docs/templates/ui/deploy-black-duck.hbs', function (assert) {
    assert.expect(1);
    assert.ok(true, 'operator-docs/templates/ui/deploy-black-duck.hbs should pass TemplateLint.\n\n');
  });
  QUnit.test('operator-docs/templates/ui/deploy-polaris.hbs', function (assert) {
    assert.expect(1);
    assert.ok(true, 'operator-docs/templates/ui/deploy-polaris.hbs should pass TemplateLint.\n\n');
  });
  QUnit.test('operator-docs/templates/ui/help.hbs', function (assert) {
    assert.expect(1);
    assert.ok(true, 'operator-docs/templates/ui/help.hbs should pass TemplateLint.\n\n');
  });
  QUnit.test('operator-docs/templates/ui/home.hbs', function (assert) {
    assert.expect(1);
    assert.ok(true, 'operator-docs/templates/ui/home.hbs should pass TemplateLint.\n\n');
  });
  QUnit.test('operator-docs/templates/ui/operator.hbs', function (assert) {
    assert.expect(1);
    assert.ok(true, 'operator-docs/templates/ui/operator.hbs should pass TemplateLint.\n\n');
  });
});
define("operator-docs/tests/lint/tests.lint-test", [], function () {
  "use strict";

  QUnit.module('ESLint | tests');
  QUnit.test('integration/components/black-duck-form-test.js', function (assert) {
    assert.expect(1);
    assert.ok(true, 'integration/components/black-duck-form-test.js should pass ESLint\n\n');
  });
  QUnit.test('integration/components/documentation-navbar-test.js', function (assert) {
    assert.expect(1);
    assert.ok(true, 'integration/components/documentation-navbar-test.js should pass ESLint\n\n');
  });
  QUnit.test('integration/components/documentation/aks-test.js', function (assert) {
    assert.expect(1);
    assert.ok(true, 'integration/components/documentation/aks-test.js should pass ESLint\n\n');
  });
  QUnit.test('integration/components/documentation/deploy-operator-test.js', function (assert) {
    assert.expect(1);
    assert.ok(true, 'integration/components/documentation/deploy-operator-test.js should pass ESLint\n\n');
  });
  QUnit.test('integration/components/documentation/eks-test.js', function (assert) {
    assert.expect(1);
    assert.ok(true, 'integration/components/documentation/eks-test.js should pass ESLint\n\n');
  });
  QUnit.test('integration/components/documentation/gke-test.js', function (assert) {
    assert.expect(1);
    assert.ok(true, 'integration/components/documentation/gke-test.js should pass ESLint\n\n');
  });
  QUnit.test('integration/components/documentation/home-test.js', function (assert) {
    assert.expect(1);
    assert.ok(true, 'integration/components/documentation/home-test.js should pass ESLint\n\n');
  });
  QUnit.test('integration/components/documentation/introduction-test.js', function (assert) {
    assert.expect(1);
    assert.ok(true, 'integration/components/documentation/introduction-test.js should pass ESLint\n\n');
  });
  QUnit.test('integration/components/documentation/on-premises-test.js', function (assert) {
    assert.expect(1);
    assert.ok(true, 'integration/components/documentation/on-premises-test.js should pass ESLint\n\n');
  });
  QUnit.test('integration/components/documentation/overview-test.js', function (assert) {
    assert.expect(1);
    assert.ok(true, 'integration/components/documentation/overview-test.js should pass ESLint\n\n');
  });
  QUnit.test('integration/components/documentation/prerequisites-test.js', function (assert) {
    assert.expect(1);
    assert.ok(true, 'integration/components/documentation/prerequisites-test.js should pass ESLint\n\n');
  });
  QUnit.test('integration/components/polaris-form-test.js', function (assert) {
    assert.expect(1);
    assert.ok(true, 'integration/components/polaris-form-test.js should pass ESLint\n\n');
  });
  QUnit.test('integration/components/ui-brand-logo-test.js', function (assert) {
    assert.expect(1);
    assert.ok(true, 'integration/components/ui-brand-logo-test.js should pass ESLint\n\n');
  });
  QUnit.test('integration/components/ui-head-test.js', function (assert) {
    assert.expect(1);
    assert.ok(true, 'integration/components/ui-head-test.js should pass ESLint\n\n');
  });
  QUnit.test('integration/components/ui-mobile-header-test.js', function (assert) {
    assert.expect(1);
    assert.ok(true, 'integration/components/ui-mobile-header-test.js should pass ESLint\n\n');
  });
  QUnit.test('integration/components/ui-nav-bar-test.js', function (assert) {
    assert.expect(1);
    assert.ok(true, 'integration/components/ui-nav-bar-test.js should pass ESLint\n\n');
  });
  QUnit.test('integration/components/ui/help-test.js', function (assert) {
    assert.expect(1);
    assert.ok(true, 'integration/components/ui/help-test.js should pass ESLint\n\n');
  });
  QUnit.test('integration/components/ui/home-test.js', function (assert) {
    assert.expect(1);
    assert.ok(true, 'integration/components/ui/home-test.js should pass ESLint\n\n');
  });
  QUnit.test('integration/components/ui/operator-test.js', function (assert) {
    assert.expect(1);
    assert.ok(true, 'integration/components/ui/operator-test.js should pass ESLint\n\n');
  });
  QUnit.test('test-helper.js', function (assert) {
    assert.expect(1);
    assert.ok(true, 'test-helper.js should pass ESLint\n\n');
  });
  QUnit.test('unit/routes/documentation-test.js', function (assert) {
    assert.expect(1);
    assert.ok(true, 'unit/routes/documentation-test.js should pass ESLint\n\n');
  });
  QUnit.test('unit/routes/documentation/aks-test.js', function (assert) {
    assert.expect(1);
    assert.ok(true, 'unit/routes/documentation/aks-test.js should pass ESLint\n\n');
  });
  QUnit.test('unit/routes/documentation/deploy-operator-test.js', function (assert) {
    assert.expect(1);
    assert.ok(true, 'unit/routes/documentation/deploy-operator-test.js should pass ESLint\n\n');
  });
  QUnit.test('unit/routes/documentation/eks-test.js', function (assert) {
    assert.expect(1);
    assert.ok(true, 'unit/routes/documentation/eks-test.js should pass ESLint\n\n');
  });
  QUnit.test('unit/routes/documentation/gke-test.js', function (assert) {
    assert.expect(1);
    assert.ok(true, 'unit/routes/documentation/gke-test.js should pass ESLint\n\n');
  });
  QUnit.test('unit/routes/documentation/home-test.js', function (assert) {
    assert.expect(1);
    assert.ok(true, 'unit/routes/documentation/home-test.js should pass ESLint\n\n');
  });
  QUnit.test('unit/routes/documentation/on-premises-test.js', function (assert) {
    assert.expect(1);
    assert.ok(true, 'unit/routes/documentation/on-premises-test.js should pass ESLint\n\n');
  });
  QUnit.test('unit/routes/documentation/overview-test.js', function (assert) {
    assert.expect(1);
    assert.ok(true, 'unit/routes/documentation/overview-test.js should pass ESLint\n\n');
  });
  QUnit.test('unit/routes/documentation/prerequisites-test.js', function (assert) {
    assert.expect(1);
    assert.ok(true, 'unit/routes/documentation/prerequisites-test.js should pass ESLint\n\n');
  });
  QUnit.test('unit/routes/ui-test.js', function (assert) {
    assert.expect(1);
    assert.ok(true, 'unit/routes/ui-test.js should pass ESLint\n\n');
  });
  QUnit.test('unit/routes/ui/deploy-black-duck-test.js', function (assert) {
    assert.expect(1);
    assert.ok(true, 'unit/routes/ui/deploy-black-duck-test.js should pass ESLint\n\n');
  });
  QUnit.test('unit/routes/ui/deploy-polaris-test.js', function (assert) {
    assert.expect(1);
    assert.ok(true, 'unit/routes/ui/deploy-polaris-test.js should pass ESLint\n\n');
  });
  QUnit.test('unit/routes/ui/help-test.js', function (assert) {
    assert.expect(1);
    assert.ok(true, 'unit/routes/ui/help-test.js should pass ESLint\n\n');
  });
  QUnit.test('unit/routes/ui/home-test.js', function (assert) {
    assert.expect(1);
    assert.ok(true, 'unit/routes/ui/home-test.js should pass ESLint\n\n');
  });
  QUnit.test('unit/routes/ui/operator-test.js', function (assert) {
    assert.expect(1);
    assert.ok(true, 'unit/routes/ui/operator-test.js should pass ESLint\n\n');
  });
});
define("operator-docs/tests/test-helper", ["operator-docs/app", "operator-docs/config/environment", "@ember/test-helpers", "ember-qunit"], function (_app, _environment, _testHelpers, _emberQunit) {
  "use strict";

  (0, _testHelpers.setApplication)(_app.default.create(_environment.default.APP));
  (0, _emberQunit.start)();
});
define("operator-docs/tests/unit/routes/documentation-test", ["qunit", "ember-qunit"], function (_qunit, _emberQunit) {
  "use strict";

  (0, _qunit.module)('Unit | Route | documentation', function (hooks) {
    (0, _emberQunit.setupTest)(hooks);
    (0, _qunit.test)('it exists', function (assert) {
      let route = this.owner.lookup('route:documentation');
      assert.ok(route);
    });
  });
});
define("operator-docs/tests/unit/routes/documentation/aks-test", ["qunit", "ember-qunit"], function (_qunit, _emberQunit) {
  "use strict";

  (0, _qunit.module)('Unit | Route | documentation/aks', function (hooks) {
    (0, _emberQunit.setupTest)(hooks);
    (0, _qunit.test)('it exists', function (assert) {
      let route = this.owner.lookup('route:documentation/aks');
      assert.ok(route);
    });
  });
});
define("operator-docs/tests/unit/routes/documentation/deploy-operator-test", ["qunit", "ember-qunit"], function (_qunit, _emberQunit) {
  "use strict";

  (0, _qunit.module)('Unit | Route | documentation/deploy-operator', function (hooks) {
    (0, _emberQunit.setupTest)(hooks);
    (0, _qunit.test)('it exists', function (assert) {
      let route = this.owner.lookup('route:documentation/deploy-operator');
      assert.ok(route);
    });
  });
});
define("operator-docs/tests/unit/routes/documentation/eks-test", ["qunit", "ember-qunit"], function (_qunit, _emberQunit) {
  "use strict";

  (0, _qunit.module)('Unit | Route | documentation/eks', function (hooks) {
    (0, _emberQunit.setupTest)(hooks);
    (0, _qunit.test)('it exists', function (assert) {
      let route = this.owner.lookup('route:documentation/eks');
      assert.ok(route);
    });
  });
});
define("operator-docs/tests/unit/routes/documentation/gke-test", ["qunit", "ember-qunit"], function (_qunit, _emberQunit) {
  "use strict";

  (0, _qunit.module)('Unit | Route | documentation/gke', function (hooks) {
    (0, _emberQunit.setupTest)(hooks);
    (0, _qunit.test)('it exists', function (assert) {
      let route = this.owner.lookup('route:documentation/gke');
      assert.ok(route);
    });
  });
});
define("operator-docs/tests/unit/routes/documentation/home-test", ["qunit", "ember-qunit"], function (_qunit, _emberQunit) {
  "use strict";

  (0, _qunit.module)('Unit | Route | documentation/home', function (hooks) {
    (0, _emberQunit.setupTest)(hooks);
    (0, _qunit.test)('it exists', function (assert) {
      let route = this.owner.lookup('route:documentation/home');
      assert.ok(route);
    });
  });
});
define("operator-docs/tests/unit/routes/documentation/on-premises-test", ["qunit", "ember-qunit"], function (_qunit, _emberQunit) {
  "use strict";

  (0, _qunit.module)('Unit | Route | documentation/on-premises', function (hooks) {
    (0, _emberQunit.setupTest)(hooks);
    (0, _qunit.test)('it exists', function (assert) {
      let route = this.owner.lookup('route:documentation/on-premises');
      assert.ok(route);
    });
  });
});
define("operator-docs/tests/unit/routes/documentation/overview-test", ["qunit", "ember-qunit"], function (_qunit, _emberQunit) {
  "use strict";

  (0, _qunit.module)('Unit | Route | documentation/overview', function (hooks) {
    (0, _emberQunit.setupTest)(hooks);
    (0, _qunit.test)('it exists', function (assert) {
      let route = this.owner.lookup('route:documentation/overview');
      assert.ok(route);
    });
  });
});
define("operator-docs/tests/unit/routes/documentation/prerequisites-test", ["qunit", "ember-qunit"], function (_qunit, _emberQunit) {
  "use strict";

  (0, _qunit.module)('Unit | Route | documentation/prerequisites', function (hooks) {
    (0, _emberQunit.setupTest)(hooks);
    (0, _qunit.test)('it exists', function (assert) {
      let route = this.owner.lookup('route:documentation/prerequisites');
      assert.ok(route);
    });
  });
});
define("operator-docs/tests/unit/routes/ui-test", ["qunit", "ember-qunit"], function (_qunit, _emberQunit) {
  "use strict";

  (0, _qunit.module)('Unit | Route | ui', function (hooks) {
    (0, _emberQunit.setupTest)(hooks);
    (0, _qunit.test)('it exists', function (assert) {
      let route = this.owner.lookup('route:ui');
      assert.ok(route);
    });
  });
});
define("operator-docs/tests/unit/routes/ui/deploy-black-duck-test", ["qunit", "ember-qunit"], function (_qunit, _emberQunit) {
  "use strict";

  (0, _qunit.module)('Unit | Route | ui/deploy_black_duck', function (hooks) {
    (0, _emberQunit.setupTest)(hooks);
    (0, _qunit.test)('it exists', function (assert) {
      let route = this.owner.lookup('route:ui/deploy-black-duck');
      assert.ok(route);
    });
  });
});
define("operator-docs/tests/unit/routes/ui/deploy-polaris-test", ["qunit", "ember-qunit"], function (_qunit, _emberQunit) {
  "use strict";

  (0, _qunit.module)('Unit | Route | ui/deploy_polaris', function (hooks) {
    (0, _emberQunit.setupTest)(hooks);
    (0, _qunit.test)('it exists', function (assert) {
      let route = this.owner.lookup('route:ui/deploy-polaris');
      assert.ok(route);
    });
  });
});
define("operator-docs/tests/unit/routes/ui/help-test", ["qunit", "ember-qunit"], function (_qunit, _emberQunit) {
  "use strict";

  (0, _qunit.module)('Unit | Route | ui/help', function (hooks) {
    (0, _emberQunit.setupTest)(hooks);
    (0, _qunit.test)('it exists', function (assert) {
      let route = this.owner.lookup('route:ui/help');
      assert.ok(route);
    });
  });
});
define("operator-docs/tests/unit/routes/ui/home-test", ["qunit", "ember-qunit"], function (_qunit, _emberQunit) {
  "use strict";

  (0, _qunit.module)('Unit | Route | ui/home', function (hooks) {
    (0, _emberQunit.setupTest)(hooks);
    (0, _qunit.test)('it exists', function (assert) {
      let route = this.owner.lookup('route:ui/home');
      assert.ok(route);
    });
  });
});
define("operator-docs/tests/unit/routes/ui/operator-test", ["qunit", "ember-qunit"], function (_qunit, _emberQunit) {
  "use strict";

  (0, _qunit.module)('Unit | Route | ui/operator', function (hooks) {
    (0, _emberQunit.setupTest)(hooks);
    (0, _qunit.test)('it exists', function (assert) {
      let route = this.owner.lookup('route:ui/operator');
      assert.ok(route);
    });
  });
});
define('operator-docs/config/environment', [], function() {
  var prefix = 'operator-docs';
try {
  var metaName = prefix + '/config/environment';
  var rawConfig = document.querySelector('meta[name="' + metaName + '"]').getAttribute('content');
  var config = JSON.parse(decodeURIComponent(rawConfig));

  var exports = { 'default': config };

  Object.defineProperty(exports, '__esModule', { value: true });

  return exports;
}
catch(err) {
  throw new Error('Could not read config from meta tag with name "' + metaName + '".');
}

});

require('operator-docs/tests/test-helper');
EmberENV.TESTS_FILE_LOADED = true;
//# sourceMappingURL=tests.map
