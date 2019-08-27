import Component from '@ember/component';

let menu_entries = [
    {
        name: "Home",
        id: "home",
        link: "/home"
    },
    {
        name: "Operator",
        id: "operator",
        link: "/operator"
    },
    {
        name: "Deploy Polaris",
        id: "deploy_polaris",
        link: "/deploy_polaris"
    },
    {
        name: "Deploy Black Duck",
        id: "deploy_black_duck",
        link: "/deploy_black_duck"
    },
    {
        name: "Docs",
        id: "docs",
        link: "/documentation"
    },
    {
        name: "Help",
        id: "help",
        link: "/help"
    }
]

export default Component.extend({
    // Variables
    menu_entries: menu_entries
});
