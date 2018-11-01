import React from 'react';
import { withStyles } from '@material-ui/core/styles';
import Table from '@material-ui/core/Table';
import TableBody from '@material-ui/core/TableBody';
import TableCell from '@material-ui/core/TableCell';
import TableHead from '@material-ui/core/TableHead';
import TableRow from '@material-ui/core/TableRow';
import Paper from '@material-ui/core/Paper';
import IconButton from '@material-ui/core/IconButton';
import DeleteIcon from '@material-ui/icons/Delete';
import JSONTree from 'react-json-tree'

const styles = theme => ({
  root: {
    marginTop: theme.spacing.unit * 3,
    overflowX: 'auto',
    margin: '0 auto',
    width: '90%'
  },
  table: {
    minWidth: 700,
  },
});

function safeHubHealth(instance) {
  try{
    return instance.hub_health.HttpStat.StartTransfer.toExponential()
  }
  catch(ex){
    return "omg"
  }
}

const InstanceTable = ({ instances, classes, handleDelete }) => {
    const namespaces = Object.keys(instances);
    if (!namespaces.length) {
        return(
            <div>
              Loading instances, hang on...
            </div>
        );
    }

    return (
        <Paper className={classes.root}>
            <Table className={classes.table}>
                <TableHead>
                    <TableRow>
                        <TableCell>Namespace</TableCell>
                        <TableCell>Size</TableCell>
                        <TableCell>Backup Interval</TableCell>
                        <TableCell>Black Duck Version</TableCell>
                        <TableCell>Database</TableCell>
                        <TableCell>PVC Claim</TableCell>
                        <TableCell>IP Address</TableCell>
                        <TableCell numeric>HubHealth</TableCell>
                        <TableCell numeric>TCP connect</TableCell>
                        <TableCell numeric></TableCell>
                    </TableRow>
                </TableHead>
                <TableBody>
                    {namespaces.map((namespace) => {
                        const instance = instances[namespace];
                        const ip = instance.status.ip ? instance.status.ip : instance.status.state;
                        const onTrashIconClick = () => {
                            return handleDelete(instance.spec.namespace);
                        };
                        // For the json tree viewer
                        const theme = {
                          scheme: 'monokai',
                          author: 'wimer hazenberg (http://www.monokai.nl)',
                          base00: '#272822',
                          base01: '#383830',
                          base02: '#49483e',
                          base03: '#75715e',
                          base04: '#a59f85',
                          base05: '#f8f8f2',
                          base06: '#f5f4f1',
                          base07: '#f9f8f5',
                          base08: '#f92672',
                          base09: '#fd971f',
                          base0A: '#f4bf75',
                          base0B: '#a6e22e',
                          base0C: '#a1efe4',
                          base0D: '#66d9ef',
                          base0E: '#ae81ff',
                          base0F: '#cc6633'
                        };
                        return (
                            <TableRow key={instance.spec.namespace}>
                                <TableCell>{instance.spec.namespace}</TableCell>
                                <TableCell>{instance.spec.flavor}</TableCell>
                                <TableCell>{instance.spec.backupInterval} {instance.spec.backupUnit}</TableCell>
                                <TableCell>{instance.spec.hubVersion}</TableCell>
                                <TableCell>{instance.spec.dbPrototype}</TableCell>
                                <TableCell>{instance.spec.pvcClaimSize ? `${instance.spec.pvcStorageClass}-${instance.spec.scanType}-${instance.spec.pvcClaimSize}` : `${instance.spec.pvcClaimSize}`}</TableCell>
                                <TableCell>
                                    {instance.status.ip ? <span><a href={`https://${instance.status.fqdn}`} target='_blank'> {instance.status.fqdn} </a><br/> <a href={`https://${ip}`} target='_blank'>{instance.status.ip}</a></span> : instance.status.state}
                                </TableCell>
                                <TableCell>
                                <JSONTree data={instance.hub_health} hideRoot={true} theme={theme} />
                                </TableCell>
                                <TableCell>{safeHubHealth(instance)}</TableCell>
                                <TableCell numeric>
                                    <IconButton aria-label="Delete instance">
                                        <DeleteIcon onClick={onTrashIconClick} />
                                    </IconButton>
                                </TableCell>
                            </TableRow>
                        );
                    })}
                </TableBody>
            </Table>
        </Paper>
    )
};

export default withStyles(styles)(InstanceTable);
