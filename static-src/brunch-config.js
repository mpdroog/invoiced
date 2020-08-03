module.exports = {
  files: {
    javascripts: {
      joinTo: {
        'vendor.js': /^(?!app)/,
        'app.js': /^app/
      }
    },
    stylesheets: {
      joinTo: 'app.css'
    }
  },

  paths: {
    public: "../static/assets"
  },

  modules: {
    autoRequire: {
      'app.js': ['main']
    }
  },

  plugins: {
    postcss: {
      processors: [
        require('csswring')()
      ]
    }
  },
  npm: {
    aliases: {
      'react': 'react-lite',
      'react-dom': 'react-lite'
    },
    globals: {
      'Big': 'big.js',
      'Moment': 'moment'
    }
  }
};
