module.exports = function(grunt) {
  grunt.loadNpmTasks('grunt-contrib-watch');
  grunt.loadNpmTasks('grunt-contrib-copy');
  grunt.loadNpmTasks('grunt-contrib-concat');
  grunt.loadNpmTasks('grunt-contrib-sass');
  grunt.loadNpmTasks('grunt-contrib-cssmin');
  grunt.loadNpmTasks('grunt-contrib-uglify');

  grunt.initConfig({
    watch: {
      sass: {
        files: ['assets/src/css/**/*.css'],
        tasks: ['concat', 'cssmin']
      },
      js: {
        files: ['assets/src/js/**/*.js'],
        tasks: ['uglify:main']
      },
    },
    concat: {
      css: {
        src: ['node_modules/normalize.css/normalize.css',
          'node_modules/font-awesome/css/font-awesome.css',
          'node_modules/animate.css/source/_base.css',
          'node_modules/animate.css/source/bouncing_entrances/bounceInRight.css',
          'node_modules/animate.css/source/fading_exits/fadeOut.css',
          'node_modules/jquery-datetimepicker/jquery.datetimepicker.css',
          'assets/src/css/main.css'
        ],
        dest: 'temp/css/main.css',
      },
    },
    copy: {
      main: {
        files: [{
          expand: true,
          flatten: true,
          src: ['node_modules/font-awesome/fonts/**'],
          dest: 'assets/fonts'
        }],
      },
    },
    cssmin: {
      options: {
        keepSpecialComments: 0
      },
      target: {
        files: [{
          expand: true,
          cwd: 'temp/css/',
          src: ['*.css', '!*.min.css'],
          dest: 'assets/css/',
          ext: '.min.css'
        }]
      }
    },
    uglify: {
      plugins: {
        files: {
          'assets/js/plugins.min.js': ['node_modules/jquery/dist/jquery.min.js',
            'node_modules/perfect-scrollbar/dist/js/min/perfect-scrollbar.jquery.min.js',
            'node_modules/showdown/dist/showdown.min.js',
            'node_modules/noty/js/noty/packaged/jquery.noty.packaged.min.js',
            'node_modules/jquery-pjax/jquery.pjax.js',
            'node_modules/jquery-serializejson/jquery.serializejson.min.js',
            'node_modules/jquery-datetimepicker/jquery.datetimepicker.js'
          ]
        }
      },
      main: {
        files: {
          'assets/js/app.min.js': ['assets/src/js/**/*.js']
        }
      }
    }
  });

  grunt.registerTask('default', ['copy', 'sass', 'concat', 'cssmin', 'uglify']);
};
