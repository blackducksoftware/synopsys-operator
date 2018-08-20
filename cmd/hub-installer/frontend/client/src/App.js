import React, { Component } from 'react';
import bdsLogo from './icon.ico';
import './App.css';
import StagingForm from './StagingForm';
import InstanceTable from './InstanceTable';
import ToastMsg from './ToastMsg';

class App extends Component {
    constructor(props) {
        super(props);
        this.state = {
            kubeSizes: [
                'small',
                'medium',
                'large',
                'OpsSight'
            ],
            backupUnits: [
                'Minute(s)',
                'Hour(s)',
                'Day(s)',
                'Week(s)'
            ],
            instances: {},
            dbInstances: [],
            pvcStorageClasses: [],
            invalidNamespace: false,
            toastMsgOpen: false,
            toastMsgText: '',
            toastMsgVariant: 'success'
        };

        this.fetchInstances = this.fetchInstances.bind(this);
        this.addInstance = this.addInstance.bind(this);
        this.removeInstance = this.removeInstance.bind(this);
        this.handleDelete = this.handleDelete.bind(this);
        this.setNamespaceStatus = this.setNamespaceStatus.bind(this);
        this.fetchDatabases = this.fetchDatabases.bind(this);
        this.fetchPVCStorageClasses = this.fetchPVCStorageClasses.bind(this);
        this.setToastStatus = this.setToastStatus.bind(this);
        this.handleToastMsgClick = this.handleToastMsgClick.bind(this);
    }

    componentDidMount() {
        this.pollInstances = setInterval(() => {
            return this.fetchInstances();
        }, 120000);
        this.fetchInstances();
        this.fetchDatabases();
        this.fetchPVCStorageClasses();
    }

    componentWillUnmount() {
        clearInterval(this.pollInstances);
    }

    //TODO: remove hardcoded tokens
    async fetchInstances() {
      console.log('Fetching customer data...');
      this.setState({
          status: "...fetching customer data, can take 20 seconds..."
      });
        const response = await fetch('/hub', {
            credentials: 'same-origin',
            headers: {
                'Content-Type': 'application/json',
                'rgb-token': 'RGB'
            },
            accept: 'application/json',
            mode: 'same-origin',
        });
        if (response.status === 200) {
            console.log('...Customer data fetched');
            const data = await response.json();
            this.setState({
                instances : data,
            });
        }
        this.setState({
            status: "Loading customers completed."
        });
    }

    async fetchDatabases() {
        const response = await fetch('/sql-instances', {
            credentials: 'same-origin',
            headers: {
                'Content-Type': 'application/json',
                'rgb-token': 'RGB'
            },
            accept: 'application/json',
            mode: 'same-origin',
        });
        if (response.status === 200) {
            console.log('DB Instances fetched');
            const dbInstances = await response.json();
            this.setState({
                dbInstances : [
                    'empty',
                    ...dbInstances
                ]
            })
        }
    }

    async fetchPVCStorageClasses() {
        const response = await fetch('/storage-classes', {
            credentials: 'same-origin',
            headers: {
                'Content-Type': 'application/json',
                'rgb-token': 'RGB'
            },
            accept: 'application/json',
            mode: 'same-origin',
        });
        if (response.status === 200) {
            console.log('Storage classes fetched');
            const pvcStorageClasses = await response.json();
            this.setState({
                pvcStorageClasses : [
                    'empty',
                    ...pvcStorageClasses
                ]
            })
        }
    }

    async handleDelete(namespace) {
        console.log("Deleting instance" + namespace);
        const response = await fetch('/hub', {
            method: 'DELETE',
            credentials: 'same-origin',
            headers: {
                'Content-Type': 'application/json',
                'rgb-token': 'RGB'
            },
            mode: 'same-origin',
            body: JSON.stringify({ namespace }),
        });

        if (response.status === 200) {
            this.setToastStatus({
                toastMsgOpen: true,
                toastMsgVariant: 'success',
                toastMsgText: 'Hub instance deleted... (Wait a few minutes for the state to be reflected in the UI).'
            });
            this.removeInstance(namespace);
            console.log('Deleted instance');
            return;
        }

        console.log(response.status);
        this.setToastStatus({
            toastMsgOpen: true,
            toastMsgVariant: 'error',
            toastMsgText: 'Hub instance not deleted, check your network settings and try again'
        });
    }

    addInstance(instance) {
        this.setState({
            instances: {
                ...this.state.instances,
                [instance.namespace] : {
                    "spec" : {
                        ...instance
                    },
                    "status":{

                    }
                }
            }
        });
        console.log(this.state.instances)
    }

    removeInstance(namespace) {
        const { [namespace] : instance, ...rest } = this.state.instances
        this.setState({
            instances: {
                ...rest
            }
        });
    }

    setNamespaceStatus(invalidNamespace) {
        if (this.state.invalidNamespace !== invalidNamespace) {
            this.setState({ invalidNamespace });
        }
    }

    setToastStatus({ toastMsgOpen, toastMsgVariant, toastMsgText }) {
        this.setState({
            toastMsgOpen,
            toastMsgVariant,
            toastMsgText
        })
    }

    handleToastMsgClick(event, reason) {
        if (reason === 'clickaway') {
            return;
        }

        this.setState({ toastMsgOpen: false });
    };

    render() {

        return (
            <div className="App">
                <header className="App-header">
                    <img src={bdsLogo} className="App-logo" alt="logo" />
                    <h1 className="App-title">Welcome to Hub Federation.</h1>
                </header>
                <StagingForm
                    kubeSizes={this.state.kubeSizes}
                    addInstance={this.addInstance}
                    setNamespaceStatus={this.setNamespaceStatus}
                    invalidNamespace={this.state.invalidNamespace}
                    dbInstances={this.state.dbInstances}
                    pvcStorageClasses={this.state.pvcStorageClasses}
                    backupUnits={this.state.backupUnits}
                    setToastStatus={this.setToastStatus}
                />

                <div className='paper-container'>
                    <InstanceTable instances={this.state.instances} handleDelete={this.handleDelete} />
                </div>
                <ToastMsg
                    message={this.state.toastMsgText}
                    variant={this.state.toastMsgVariant}
                    toastMsgOpen={this.state.toastMsgOpen}
                    onClose={this.handleToastMsgClick}
                />
            </div>
        );
    }
}

export default App;
