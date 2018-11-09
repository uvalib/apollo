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

    var fixedElement = $(this.id)
    if (fixedElement.length === 0 ) return

    // before anyhting is changed, record the original top of
    // the fixed item in a data property so its position 
    // can be restored
    let origVal = fixedElement.data("origTop")
    if ( !origVal ) {
      let ot = fixedElement.offset().top
      fixedElement.data("origTop", ot)
    }

    // if the fixed item is scrolled where it would go offscreen,
    // reset its top top keep it at the top of the display
    let scrollTop= $(window).scrollTop();
    if ( scrollTop >= this.pinTop ) {
       fixedElement.offset({top: scrollTop+offset});
    } else {
       fixedElement.offset({top: fixedElement.data("origTop")});
    }
  }
}
