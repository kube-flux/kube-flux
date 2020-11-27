import React, { useState, useEffect, useRef } from 'react';
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

function createData(name, num) {
  return { name, num };
}

export default function BasicTable({ currentStatus }) {
  const classes = useStyles();
  const [rows, setRows] = useState([])

  function capitalize(word) {
    return word.charAt(0).toUpperCase() + word.slice(1);
  }

  function useInterval(callback, delay) {
    const savedCallback = useRef();
  
    // Remember the latest callback.
    useEffect(() => {
      savedCallback.current = callback;
    }, [callback]);
  
    // Set up the interval.
    useEffect(() => {
      function tick() {
        savedCallback.current();
      }
      if (delay !== null) {
        let id = setInterval(tick, delay);
        return () => clearInterval(id);
      }
    }, [delay]);
  }

  useInterval(() => {
      const headers = { 'Content-Type': 'application/json' }
      fetch('http://localhost:8888/policy', headers)
        .then(response => response.json())
        .then(data => {
            let currStatusCap = capitalize(currentStatus)
            console.log(currStatusCap)

            let result = data.Factor[currStatusCap]
            let rowsArr = []

            for (const property in result) {
                rowsArr.push(createData(property, result[property]))
            }

            setRows(rowsArr)
        })
  }, 1000)

  return (
    <TableContainer component={Paper}>
      <Table className={classes.table} aria-label="simple table">
        <TableHead>
          <TableRow>
            <TableCell>Deployments</TableCell>
            <TableCell align="right">Number of Pods</TableCell>
          </TableRow>
        </TableHead>
        <TableBody>
          {rows.map((row) => (
            <TableRow key={row.name}>
              <TableCell component="th" scope="row">
                {row.name}
              </TableCell>
              <TableCell align="right">{row.num}</TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </TableContainer>
  );
}