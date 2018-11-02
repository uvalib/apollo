// PinnedScroll is used to pin a dom element to a particular location when scrolling.
// It can be used to stop a header or toolbar from scrolling offscreen. Pass  the
// jQuery selector of the raeget element and the window scroll top at which  to pin the element.
// Optionally pass an offset that can be used if the pinned item sits below another pinned item.
export default class PinnedScroll {
  constructor( id, pinTop, offsetId) {
    this.id = id
    this.pinTop = pinTop
    this.offsetId = offsetId
  }

  register() {
    window.addEventListener("scroll", this.handleScroll.bind(this))
  }

  unregister() {
    window.removeEventListener("scroll", this.handleScroll)
  }

  handleScroll( ) {
    let offset = 0
    if (this.offsetId) {
      offset = $(this.offsetId).outerHeight(true)
    }
    var fixedHeader = $(this.id)
    if (fixedHeader.length === 0 ) return
    let origVal = fixedHeader.data("origTop")
    if ( !origVal ) {
      let ot = fixedHeader.offset().top
      fixedHeader.data("origTop", ot)
    }
    let scrollTop= $(window).scrollTop();
    if ( scrollTop >= this.pinTop ) {
       fixedHeader.offset({top: scrollTop+offset});
    } else {
       fixedHeader.offset({top: fixedHeader.data("origTop")});
    }
  }
}
