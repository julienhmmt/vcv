import { mount } from 'svelte'
import './app.css'
import Admin from './Admin.svelte'

const app = mount(Admin, {
  target: document.getElementById('app')!,
})

export default app
