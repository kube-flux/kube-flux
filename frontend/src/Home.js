import React, { useState, useEffect } from 'react';
import './App.css';
import NavBar from './components/NavBar'
import Table from './components/Table'
import { Button, Card, Divider, Icon, Header } from 'semantic-ui-react'

export default function Home() {
    
    const [currStatus, setStatus] = useState()

    useEffect(() => {
        const headers = { 'Content-Type': 'application/json' }
        fetch('http://localhost:8888/policy', headers)
            .then(response => response.json())
            .then(data => setStatus(data.Status.toLowerCase()))
    })

    function handleOnClick(status) {
        const data = { "Status": status, "Factor": "null" }
        if (status.toLowerCase() !== currStatus) {
            fetch('http://localhost:8888/policy', {
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(data)
            })
            .then(response => setStatus(status.toLowerCase()))
        }
    }

    return (
        <div>
            <div>
            <NavBar/>
            </div>

            <div>
                <br></br>
                <Card header='Current Status' color={currStatus} description={currStatus}></Card>
            </div>

            <div>
                <Divider horizontal>
                    <Header as='h4'>
                        <Icon name='pencil alternate' />
                        Select an input
                    </Header>
                </Divider>
                <Button basic color='green' size='huge' content='Green' onClick={() => handleOnClick('Green')}/>
                {' '}
                <Button basic color='yellow' size='huge' content='Yellow' onClick={() => handleOnClick('Yellow')} />
                {' '}
                <Button basic color='red' size='huge' content='Red' onClick={() => handleOnClick('Red')}/>
            </div>
            
            <div>
                <br></br>
                <Divider horizontal>
                    <Header as='h4'>
                        <Icon name='list' />
                        Deployments
                    </Header>
                </Divider>
                
                <Table currentStatus={currStatus}/>
            </div>   
        </div>
  );
}