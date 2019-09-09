import EmberRouter from '@ember/routing/router';
import config from './config/environment';

const Router = EmberRouter.extend({
  location: config.locationType,
  rootURL: config.rootURL
});

Router.map(function() {
  this.route('documentation', function() {
    this.route('gke');
    this.route('eks');
    this.route('aks');
    this.route('on-premises');
    this.route('prerequisites');
    this.route('home');
    this.route('deploy-operator');
  });
  this.route('ui');
  this.route('ui', { path: '/' }, function() {
    this.route('home');
    this.route('help');
    this.route('operator');
    this.route('deploy_polaris');
    this.route('deploy_black_duck');
  });
});

export default Router;
