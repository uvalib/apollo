<template>
   <footer>
   <div class="address">
       <a href="http://www.library.virginia.edu/">University of Virginia Library</a><br>
       PO Box 400113, Charlottesville, VA 22904-4113<br>
       ph: (434) 924-3021, fax: (434) 924-1431, <a href="mailto:library@virginia.edu">library@virginia.edu</a>
   </div>
   <div class="external_links">
       <a href="http://www.library.virginia.edu/staff/">Staff Directory</a> | <a href="http://www.virginia.edu/">U.Va. Home</a> | <a href="http://its.virginia.edu/">ITS</a><br>
       <a href="mailto:site-feedback@virginia.edu?subject=U.Va.%20Library%20Site%20Feedback&amp;body=Thank%20you%20for%20your%20feedback!%20%20Are%20you%20having%20trouble%20finding%20information%20on%20the%20new%20Library%20site%20or%20have%20you%20encountered%20an%20error?%20Please%20let%20us%20know%20below.%0A%0APage%20you%20were%20on%3A%20https://tracksys.lib.virginia.edu/%0A%0AMy%20feedback%3A%20">Feedback</a>
       | <a href="http://www.library.virginia.edu/askalibrarian/">Questions? Ask a Librarian</a>
       <br>
       <a href="http://www.library.virginia.edu/site-search/">Site Search</a> | <a href="http://www.library.virginia.edu/hours/">Hours</a> | <a href="http://www.library.virginia.edu/map/">Map</a> | <a href="http://www.library.virginia.edu/policies/">Policies</a> | <a href="http://www.library.virginia.edu/press/">Press</a> | <a href="http://www.library.virginia.edu/jobs/">Jobs</a><br>
       <a href="http://www2.lib.virginia.edu/campaign/">Give to the Library</a>
       <p class="copy">
          © 2017 by the Rector and Visitors of the <a href="http://www.virginia.edu/">University of Virginia</a>
       </p>
    </div>
    <div class="legal">
       <p class="gov"><a class="fdlp" href="http://guides.lib.virginia.edu/findinggovinfo"><span id="fdlp-badge"></span></a> This library is a Congressionally designated depository for <a href="http://guides.lib.virginia.edu/findinggovinfo">U.S. Government documents</a>. Public access to the Government documents is guaranteed by public law.</p>
    </div>
   </footer>
</template>

<script>
  export default {
    name: 'apollo-footer'
  }
  </script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>
  footer {
     border-top:1px solid #d2e5f7;
     background-color: #002F6C;
     color: #d2e5f7;
     padding:20px;
     position:relative;
     font-size: 0.9em;
  }

  footer .external_links {
     float:right;
     text-align: right;
  }
  footer p.copy {
     margin-top:20px;
  }
  footer a {
     color: #d2e5f7;
  }
  footer a:hover {
     color: white;
  }
  footer a.fdlp {
     border-bottom: none !important;
  }
  footer div.address {
     float:left;
  }
  footer div.legal {
     position: relative;
     text-align: center;
     clear: both;
  }
  footer #fdlp-badge {
     background: url(../assets/fdlp-badge.png);
     height: 34px;  width: 39px;
     display: inline-block;
     position: relative;
     top: 8px;
     left: -8px;
  }
  footer p.copy, footer p.gov {
     vertical-align: middle;
  }
</style>
