import Route from '@ember/routing/route';

export default Route.extend({
    model() {
        return [{
            first: "black",
            last: "duck"
        },
        {
            first: "white",
            last: "duck"
        }];
    }
});
