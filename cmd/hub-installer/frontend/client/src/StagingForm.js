import React, { Component } from 'react';
import PropTypes from 'prop-types';
import classnames from 'classnames';
import { withStyles } from '@material-ui/core/styles';
import MenuItem from '@material-ui/core/MenuItem';
import TextField from '@material-ui/core/TextField';
import Button from '@material-ui/core/Button';
import Radio from '@material-ui/core/Radio';
import RadioGroup from '@material-ui/core/RadioGroup';
import FormControlLabel from '@material-ui/core/FormControlLabel';
import FormControl from '@material-ui/core/FormControl';
import FormLabel from '@material-ui/core/FormLabel';
// import deepPurple from '@material-ui/core/colors/purple';

//TODO: figure out child selectors/dynamic styles
const styles = theme => ({
    container: {
        display: 'flex',
        flexWrap: 'wrap',
    },
    formContainer: {
        margin: '0 auto',
        width: '80%'
    },
    textField: {
        marginLeft: theme.spacing.unit,
        marginRight: theme.spacing.unit,
        width: 500,
    },
    singleRowFields: {
        display: 'flex'
    },
    singleRowFieldLeft: {
        marginRight: theme.spacing.unit / 2,
        flex: 1
    },
    singleRowFieldRight: {
        marginLeft: theme.spacing.unit / 2,
        flex: 1
    },
    menu: {
        width: 200,
    },
    button: {
        margin: theme.spacing.unit,
    },
    rightIcon: {
        marginLeft: theme.spacing.unit,
    },
    formControl: {
        margin: theme.spacing.unit * 3,
    },
    group: {
        margin: `${theme.spacing.unit}px 0`,
        flexDirection: 'row'
    },
    close: {
        width: theme.spacing.unit * 4,
        height: theme.spacing.unit * 4,
    },
});

const initialState = {
    namespace: '',
    flavor: 'small',
    hubTimeout: '2',
    dockerRegistry: 'gcr.io',
    dockerRepo: 'gke-verification/blackducksoftware',
    hubVersion: '4.7.0',
    dbPrototype: 'empty',
    status: 'pending',
    token: '',
    emptyFormFields: true
};

class StagingForm extends Component {
    constructor(props) {
        super(props);
        this.state = initialState;

        // TODO: React docs - transform pkg, don't need to bind
        this.handleChange = this.handleChange.bind(this);
        this.handleSubmit = this.handleSubmit.bind(this);
        this.resetForm = this.resetForm.bind(this);
        this.validateNamespace = this.validateNamespace.bind(this);
        this.emptyFormFields = this.emptyFormFields.bind(this);
    }

    componentDidMount() {
        this.namespaceField.addEventListener('blur', this.validateNamespace);
    }

    componentWillUnmount() {
        this.namespaceField.removeEventListener('blur', this.validateNamespace)
    }

    handleChange(event) {
        const stateKey = event.target.name;
        this.setState({ [stateKey]: event.target.value }, () => {
            this.emptyFormFields();
        });
    }

    resetForm() {
        this.setState(initialState)
    }

    async handleSubmit(event) {
        event.preventDefault();
        const {
            token,
            emptyFormFields,
            dbPrototype,
            ...formData
        } = this.state;
        const database = dbPrototype === 'empty' ? '' : dbPrototype;
        const response = await fetch('/hub', {
            method: 'POST',
            credentials: 'same-origin',
            headers: {
                'Content-Type': 'application/json',
                'rgb-token': token
            },
            mode: 'same-origin',
            body: JSON.stringify({ ...formData, dbPrototype: database }),
        });

        if (response.status === 200) {
            this.props.setToastStatus({
                toastMsgOpen: true,
                toastMsgVariant: 'success',
                toastMsgText: 'Hub instance submitted! IP address will appear shortly'
            });
            this.props.addInstance(formData);
            this.resetForm()
            return;
        }

        this.props.setToastStatus({
            toastMsgOpen: true,
            toastMsgVariant: 'error',
            toastMsgText: 'Invalid token, check your token and try again ( error code '+response.status+"')'"
        });
    }

    validateNamespace(event) {
        const regExp = RegExp(/^[a-z0-9]([-a-z0-9]*[a-z0-9])?$/);
        const invalidNamespace = !regExp.test(event.target.value);
        this.props.setNamespaceStatus(invalidNamespace);
    }

    emptyFormFields() {
        const {
            flavor,
            hubTimeout,
            status,
            emptyFormFields : emptyFields,
            ...textFields
        } = this.state;
        const emptyFormFields = Object.keys(textFields).some((field) => !Boolean(textFields[field]));
        if (emptyFormFields !== this.state.emptyFormFields) {
            this.setState({ emptyFormFields });
        }
    }

    render() {
        const {
            classes,
            invalidNamespace,
            kubeSizes,
            dbInstances
        } = this.props;
        // const primary = deepPurple[200];

        return (
            <div className={classes.formContainer}>
                <form
                    id="staging-form"
                    className={classes.container}
                    noValidate
                    autoComplete="off"
                >
                    <TextField
                        id="namespace"
                        name="namespace"
                        label="Namespace"
                        className={classes.textField}
                        value={this.state.namespace}
                        onChange={this.handleChange}
                        margin="normal"
                        autoFocus
                        inputRef={el => this.namespaceField = el}
                        error={invalidNamespace}
                        helperText="Lowercase letters, numbers, and hyphens only. Cannot start or end with hypens."
                    />
                    <div className={classes.root}>
                        <FormControl component="fieldset" className={classes.formControl}>
                            <FormLabel component="legend">HUB Size</FormLabel>
                            <RadioGroup
                                aria-label="HUB Size"
                                name="flavor"
                                className={classes.group}
                                value={this.state.flavor}
                                onChange={this.handleChange}
                            >
                                {kubeSizes.map((size) => {
                                    return (
                                        <FormControlLabel
                                            key={`flavor-${size}`}
                                            value={size}
                                            control={<Radio color="primary" />}
                                            label={size}
                                        />
                                    );
                                })}
                            </RadioGroup>
                        </FormControl>
                    </div>
                    <div className={classnames(classes.singleRowFields, classes.textField)}>
                        <TextField
                            id="hubVersion"
                            name="hubVersion"
                            label="Hub Version"
                            className={classes.singleRowFieldRight}
                            value={this.state.hubVersion}
                            onChange={this.handleChange}
                            margin="normal"
                        />
                    </div>
                    <TextField
                        id="dockerRegistry"
                        name="dockerRegistry"
                        label="Docker Registry"
                        className={classes.textField}
                        value={this.state.dockerRegistry}
                        onChange={this.handleChange}
                        margin="normal"
                    />
                    <TextField
                        id="dockerRepo"
                        name="dockerRepo"
                        label="Docker Repo"
                        className={classes.textField}
                        value={this.state.dockerRepo}
                        onChange={this.handleChange}
                        margin="normal"
                    />
                    <TextField
                        select
                        id="dbPrototype"
                        name="dbPrototype"
                        label="Database"
                        className={classes.textField}
                        value={this.state.dbPrototype}
                        onChange={this.handleChange}
                        SelectProps={{
                            MenuProps: {
                                className: classes.menu,
                            },
                        }}
                        margin="normal"
                    >
                        {dbInstances.map((instance) => {
                            return (
                                <MenuItem key={`instance-${instance}`} value={instance}>
                                    {instance}
                                </MenuItem>
                            );
                        })}
                    </TextField>
                    <TextField
                        id="token"
                        name="token"
                        label="Token"
                        className={classes.textField}
                        value={this.state.token}
                        onChange={this.handleChange}
                        margin="normal"
                    />
                    <Button
                        variant="contained"
                        size="medium"
                        className={classes.button}
                        type='submit'
                        color="primary"
                        onClick={this.handleSubmit}
                        disabled={this.state.emptyFormFields || invalidNamespace}
                    >
                        Submit
                    </Button>
                </form>
            </div>
        );
    }
}

export default withStyles(styles)(StagingForm);

StagingForm.propTypes = {
    addInstance: PropTypes.func,
    dbInstances: PropTypes.arrayOf(PropTypes.string),
    invalidNamespace: PropTypes.bool,
    kubeSizes: PropTypes.arrayOf(PropTypes.string)
}
