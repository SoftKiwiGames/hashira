"use strict";

class HashiraEditor {
    constructor(hashira) {
        this.hashira = hashira;
        this.dragStartX = 0;
        this.dragStartY = 0;
        this.dragging = false;
    }

    bindEvents = (canvas) => {
        canvas.addEventListener("mousedown", this.onCanvasMouseDown);
        canvas.addEventListener("mouseup", this.onCanvasMouseUp);
        canvas.addEventListener("mousemove", this.onCanvasMouseMove);
        canvas.addEventListener("mousemove", this.onCanvasMouseMove);
        canvas.addEventListener("contextmenu", e => e.preventDefault());
        canvas.addEventListener("wheel", this.onCanvasWheel);
    }

    onCanvasMouseDown = (e) => {
        e.preventDefault();
        if (this._isRightMB(e)) {
            this.dragStartX = e.clientX;
            this.dragStartY = e.clientY;
            this.dragging = true;
        }
    }

    onCanvasMouseUp = (e) => {
        e.preventDefault();
        this.dragging = false;
    }

    onCanvasMouseMove = (e) => {
        e.preventDefault();
        if (this.dragging) {
            const x = e.clientX;
            const y = e.clientY;

            const dx = x - this.dragStartX;
            const dy = y - this.dragStartY;

            this.dragStartX = x;
            this.dragStartY = y;

            this._rightMBDraggedBy(dx, dy);
        }
    }

    onCanvasWheel = (e) => {
        e.preventDefault();
        const by = Math.sign(e.deltaY);
        this.hashira.setCameraZoomBy(-by)
    }

    _rightMBDraggedBy = (dx, dy) => {
        this.hashira.setCameraTranslationBy(-dx, dy);
    }

    _isRightMB = (e) => {
        let isRightMB = false;
        if ("which" in e)  // Gecko (Firefox), WebKit (Safari/Chrome) & Opera
            isRightMB = e.which == 3;
        else if ("button" in e)  // IE, Opera 
            isRightMB = e.button == 2;
        return isRightMB;
    }
}