import Component from '@ember/component';

let menu_entries = [
    {
        name: "Home",
        id: "home",
        link: "/documentation/home"
    },
    {
        name: "Prerequisites",
        id: "prerequisites",
        link: "/documentation/prerequisites"
    },
    {
        name: "Installing or Upgrading Synopsys Operator",
        id: "deploy-operator",
        link: "/documentation/deploy-operator"
    },
    {
        name: "Deploy Polaris on On-Premises Kubernetes",
        id: "on-premises",
        link: "/documentation/on-premises"
    },
    {
        name: "Deploy Polaris on Google Kubernetes Engine(GKE)",
        id: "gke",
        link: "/documentation/gke"
    },
    {
        name: "Deploy Polaris on Elastic Kubernetes Service(EKS)",
        id: "eks",
        link: "/documentation/eks"
    },
    {
        name: "Deploy Polaris on Azure Kubernetes Service(AKS)",
        id: "aks",
        link: "/documentation/aks"
    },
    {
        name: "Contact",
        id: "contact",
        link: "/documentation/contact"
    }
]

export default Component.extend({
    // Variables
    menu_entries: menu_entries
});
