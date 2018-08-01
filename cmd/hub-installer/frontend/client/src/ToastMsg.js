import React from 'react';
import PropTypes from 'prop-types';
import classNames from 'classnames';
import CheckCircleIcon from '@material-ui/icons/CheckCircle';
import ErrorIcon from '@material-ui/icons/Error';
import CloseIcon from '@material-ui/icons/Close';
import green from '@material-ui/core/colors/green';
import IconButton from '@material-ui/core/IconButton';
import Snackbar from '@material-ui/core/Snackbar';
import SnackbarContent from '@material-ui/core/SnackbarContent';
import { withStyles } from '@material-ui/core/styles';

const variantIcon = {
  success: CheckCircleIcon,
  error: ErrorIcon
};

const styles = theme => ({
  success: {
    backgroundColor: green[600],
  },
  error: {
    backgroundColor: theme.palette.error.dark,
  },
  icon: {
    fontSize: 20,
  },
  iconVariant: {
    opacity: 0.9,
    marginRight: theme.spacing.unit,
  },
  message: {
    display: 'flex',
    alignItems: 'center',
  },
});


const ToastMsg = (props) => {
    const {
        classes,
        className,
        message,
        onClose,
        toastMsgOpen,
        variant,
        ...other
    } = props;
    const Icon = variantIcon[variant];

    return (
        <Snackbar
          anchorOrigin={{
            vertical: 'top',
            horizontal: 'center',
          }}
          open={toastMsgOpen}
          autoHideDuration={6000}
          onClose={this.handleClose}
        >
            <SnackbarContent
                className={classNames(classes[variant], className)}
                aria-describedby="toast-msg"
                message={
                    <span id="toast-msg" className={classes.message}>
                        <Icon className={classNames(classes.icon, classes.iconVariant)} />
                        {message}
                    </span>
                }
                action={[
                    <IconButton
                        key="close"
                        aria-label="Close"
                        color="inherit"
                        className={classes.close}
                        onClick={onClose}
                    >
                        <CloseIcon className={classes.icon} />
                    </IconButton>,
                ]}
                {...other}
            />
        </Snackbar>
    )
}

export default withStyles(styles)(ToastMsg);

ToastMsg.propTypes = {
    classes: PropTypes.object.isRequired,
    className: PropTypes.string,
    message: PropTypes.node,
    onClose: PropTypes.func,
    toastMsgOpen: PropTypes.bool,
    variant: PropTypes.oneOf(['success', 'warning', 'error', 'info']).isRequired
}
