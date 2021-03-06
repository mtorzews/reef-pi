import React from 'react'
import Enzyme, { shallow } from 'enzyme'
import Adapter from 'enzyme-adapter-react-16'
import App from './app'
import configureMockStore from 'redux-mock-store'
import renderer from 'react-test-renderer'
import thunk from 'redux-thunk'
import { Provider } from 'react-redux'
import 'isomorphic-fetch'
import JackSelector from './jack_selector'
import SelectEquipment from './select_equipment'
import SignIn from './sign_in'
import fetchMock from 'fetch-mock'

Enzyme.configure({ adapter: new Adapter() })
const mockStore = configureMockStore([thunk])

describe('App', () => {
  afterEach(() => {
    fetchMock.reset()
    fetchMock.restore()
  })
  it('<App />', () => {
    SignIn.isSignedIn = jest.fn().mockImplementation(() => {
      return new Promise(function (resolve) {
        return resolve(true)
      })
    })
    const m = shallow(<App store={mockStore({ info: {}, capabilities: [] })} />).instance()
    m.state.loaded = true
    m.render()
    m.state.logged = true
    m.getComponent()
    m.componentDidMount()
  })

  it('<JackSelector />', () => {
    const jacks = [{ id: '1', name: 'Foo', pins: [1, 2] }]
    const m = shallow(<JackSelector store={mockStore({ jacks: jacks })} id='1' update={() => {}} />)
      .dive()
      .instance()
    m.setJack(0)
  })

  it('<SelectEquipment />', () => {
    const eqs = [{ id: '1', name: 'foo' }]
    const m = shallow(<SelectEquipment store={mockStore({ equipment: eqs })} active='1' update={() => {}} />)
      .dive()
      .instance()
    m.setEquipment(0)()
    m.setEquipment('none')()
    shallow(<SelectEquipment store={mockStore({ equipment: eqs })} readOnly active='1' update={() => {}} />)
      .dive()
      .instance()
  })
})
