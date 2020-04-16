import resolve from '@rollup/plugin-node-resolve'
import commonjs from '@rollup/plugin-commonjs'

export default {
  input: 'assets/main.js',
  output: {
    file: 'public/app.js',
    format: 'iife'
  },
  plugins: [ resolve(), commonjs() ]
}
