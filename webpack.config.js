const webpack = require('webpack');

module.exports = {
  entry: './static/js/src/index.js',
  output: {
    path: './static/js/dist/',
    filename: 'bundle.js',
  },
  module: {
    loaders: [
      {
        test: /\.(js|jsx)$/,
        loader: 'babel',
        exclude: /node_modules/,
        query: {
          presets: ['es2015'],
        },
      },
    ],
  },
  resolve: {
    extensions: ['', '.js', '.json'],
  },
};
