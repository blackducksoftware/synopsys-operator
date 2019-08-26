import Component from '@ember/component';

let menu_entries = [
    {
        name: "Home",
        id: "home",
        link: "/documentation/home"
    },
    {
        name: "Polaris Requirements",
        id: "prerequisites",
        link: "/documentation/prerequisites"
    },
    {
        name: "Instructions to deploy Polaris operator",
        id: "install-operator",
        link: "/documentation/install-operator"
    },
    {
        name: "Instructions to deploy Polaris on On-Premises Kubernetes",
        id: "on-premises",
        link: "/documentation/on-premises"
    },
    {
        name: "Instructions to deploy Polaris on Google Kubernetes Engine(GKE)",
        id: "gke",
        link: "/documentation/gke"
    },
    {
        name: "Instructions to deploy Polaris on Elastic Kubernetes Service(EKS)",
        id: "eks",
        link: "/documentation/eks"
    },
    {
        name: "Instructions to deploy Polaris on Azure Kubernetes Service(AKS)",
        id: "aks",
        link: "/documentation/aks"
    },
    {
        name: "Additional Help",
        id: "contact",
        link: "/documentation/contact"
    }
]

export default Component.extend({
    menu_entries: menu_entries
});
