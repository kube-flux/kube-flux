import React from 'react';
import { Menu, Segment, Icon } from 'semantic-ui-react'

export default function NavBar() {
    return (
        <Segment inverted>
          <Menu inverted pointing secondary icon='labeled' size='large'>
            <Menu.Item
              name='Kube-Flux'
              active='true'
            >
                <Icon name='hdd outline'/>
                Kube-Flux
            </Menu.Item>
          </Menu>
        </Segment>
      )
}