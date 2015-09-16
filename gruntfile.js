module.exports = function(grunt) {

  grunt.loadNpmTasks('grunt-contrib-watch');
  grunt.loadNpmTasks('grunt-contrib-copy');
  grunt.loadNpmTasks('grunt-contrib-concat');
  grunt.loadNpmTasks('grunt-contrib-sass');
  grunt.loadNpmTasks('grunt-contrib-cssmin');
  grunt.loadNpmTasks('grunt-contrib-uglify');

  grunt.initConfig({
    watch: {
      files: [
        'assets/src/js/**/*.js',
        'assets/src/css/sass/**/*.scss'
      ],
      tasks: ['sass', 'concat', 'cssmin', 'uglify']
    },
    sass: {
      dist: {
        options: {
          style: 'expanded',
          sourcemap: 'none'
        },
        files: [{
          expand: true,
          cwd: 'assets/css/src/sass',
          src: ['**/*.scss'],
          dest: 'assets/css/src',
          ext: '.css'
        }]
      }
    },
    concat: {
      css: {
        src: ['node_modules/normalize.css/normalize.css',
          'node_modules/font-awesome/css/font-awesome.css',
          'assets/src/css/main.css'
        ],
        dest: 'assets/css/src/main.css',
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
      target: {
        files: [{
          expand: true,
          cwd: 'assets/css/src',
          src: ['*.css', '!*.min.css'],
          dest: 'assets/css/',
          ext: '.min.css'
        }]
      }
    },
    uglify: {
      target: {
        files: {
          'assets/js/app.min.js': ['node_modules/jquery/dist/jquery.js',
            'node_modules/perfect-scrollbar/dist/js/perfect-scrollbar.jquery.js',
            'node_modules/showdown/dist/showdown.js',
            'assets/src/js/**/*.js'
          ]
        }
      }
    }
  });

  grunt.registerTask('default', ['copy', 'sass', 'concat', 'cssmin', 'uglify']);
};