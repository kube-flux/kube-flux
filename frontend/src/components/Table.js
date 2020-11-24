import React from 'react';
import { makeStyles } from '@material-ui/core/styles';
import Table from '@material-ui/core/Table';
import TableBody from '@material-ui/core/TableBody';
import TableCell from '@material-ui/core/TableCell';
import TableContainer from '@material-ui/core/TableContainer';
import TableHead from '@material-ui/core/TableHead';
import TableRow from '@material-ui/core/TableRow';
import Paper from '@material-ui/core/Paper';

const useStyles = makeStyles({
  table: {
    minWidth: 650,
  },
});

function createData(name, low, medium, top) {
  return { name, low, medium, top};
}

const rows = [
  createData('hello-world1', 4, 4, 4),
  createData('hello-world2', 0, 4, 0),
  createData('hello-world3', 4, 2, 4),
  createData('hello-world4', 0, 4, 4),
];

export default function BasicTable() {
  const classes = useStyles();

  return (
    <TableContainer component={Paper}>
      <Table className={classes.table} aria-label="simple table">
        <TableHead>
          <TableRow>
            <TableCell>Deployments</TableCell>
            <TableCell align="right">Low</TableCell>
            <TableCell align="right">Medium</TableCell>
            <TableCell align="right">Top</TableCell>
          </TableRow>
        </TableHead>
        <TableBody>
          {rows.map((row) => (
            <TableRow key={row.name}>
              <TableCell component="th" scope="row">
                {row.name}
              </TableCell>
              <TableCell align="right">{row.low}</TableCell>
              <TableCell align="right">{row.medium}</TableCell>
              <TableCell align="right">{row.top}</TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </TableContainer>
  );
}