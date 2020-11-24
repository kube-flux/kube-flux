import React from 'react';
import './App.css';
import NavBar from './components/NavBar'
import Table from './components/Table'
import { Button, Card, Divider, Icon, Header, Message } from 'semantic-ui-react'
import { withStyles, makeStyles } from '@material-ui/core/styles';
import IframeResizer from 'iframe-resizer-react'

const useStyles = makeStyles((theme) => ({
	root: {
		width: 300 + theme.spacing(3) * 2,
	},
	margin: {
		height: theme.spacing(3),
	},
	table: {
		minWidth: 650,
	},
}));

export default function Home() {
    const classes = useStyles();
    return (
        <div>
        
        <div>
        <NavBar/>
        </div>

        <div>
            <br></br>
            <Card header='Current Status' color='green' description='Green'></Card>
        </div>

        <div>
            <Divider horizontal>
                <Header as='h4'>
                    <Icon name='pencil alternate' />
                    Select an input
                </Header>
            </Divider>
            <Button basic color='green' size='huge' content='Green' />
            {' '}
            <Button basic color='yellow' size='huge' content='Yellow' />
            {' '}
            <Button basic color='red' size='huge' content='Red'/>
        </div>
        
        <div>
            <Table/>
        </div>   
    </div>
  );
}