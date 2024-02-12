// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0
var gridClass = ".grid";
var localStorageKey = "gridState";
var grid = initGrid(layout);

// Editable canvas is used to make the canvas editable allowing the user to add widgets and be able to move the
// widgets around the canvas
function editableCanvas() {
  grid = editGrid(grid, layout);
}

function saveCanvas() {
  saveGrid(grid, dashboardID);
}

function cancelEdit() {
  cancelEditGrid(grid);
}
// Config has the ID, Content and Script parameters
function addWidget(config) {
  // Create a new grid item
  const newItem = document.createElement("div");
  newItem.className = "item";
  newItem.classList.add("item-editable");
  newItem.innerHTML = `
    <button type="button" class="btn btn-sm" id="removeItem" onclick="removeGridItem(this.parentNode);">
      <i class="fas fa-trash-can"></i>
    </button>
    <div class="item-content" id="${config.ID}" style="width: 500px;height:400px;">
      ${config.Content}
    </div>
  `;
  if (config.Script) {
    var scriptTag = document.createElement("script");
    scriptTag.type = "text/javascript";
    scriptTag.defer = true;
    scriptTag.innerHTML = config.Script;
    newItem.appendChild(scriptTag);
  }
  grid.add(newItem);
  resizeObserver.observe(newItem);
}

function removeGridItem(item) {
  grid.remove(grid.getItems(item), { removeElements: true });
}

function openWidgetModal(widget) {
  const widgetModal = new bootstrap.Modal(document.getElementById(`${widget}Modal`));
  widgetModal.show();
}

function initGrid(layout) {
  if (layout) {
    loadLayout(layout);
  } else {
    showNoWidgetPlaceholder();
  }

  return grid;
}

function saveLayout(grid, dashboardID) {
  const itemData = grid.getItems().map((item) => {
    const hasWidgetScript =
      item._element.children[2] && item._element.children[2].innerHTML.trim() !== "";
    return {
      innerHTML: item._element.innerHTML,
      widgetID: item._element.children[1].children[0].id,
      ...(hasWidgetScript ? { widgetScript: item._element.children[2].innerHTML } : {}),
    };
  });

  const gridState = {
    items: itemData,
    layout: grid._layout,
    settings: {
      dragEnabled: grid._settings.dragEnabled,
      // Add other relevant settings if needed
    },
  };

  // Convert the gridState to a JSON string
  const jsonString = JSON.stringify(gridState, function (key, value) {
    // Exclude circular references
    if (key === "_item" || key === "_grid" || key === "_layout") {
      return undefined;
    }
    return value;
  });

  const dashboard = {
    id: dashboardID,
    layout: jsonString,
  };

  fetch("/dashboards", {
    method: "PATCH",
    body: JSON.stringify(dashboard),
    headers: {
      "Content-Type": "application/json",
    },
  }).then((response) => {
    if (!response.ok) {
      const errorMessage = response.headers.get("X-Error-Message");
      console.log("Error: ", errorMessage);
    } else {
      window.location.reload();
    }
  });
}

function loadLayout(savedLayout) {
  try {
    const gridState = JSON.parse(savedLayout);
    // Clear the existing grid
    if (grid) {
      grid.destroy(true);
    }

    grid = new Muuri(gridClass, {
      dragEnabled: gridState.settings.dragEnabled,
      dragHandle: ".item-content",
      // Add any other relevant settings
    });

    if (gridState.items.length === 0) {
      showNoWidgetPlaceholder();
    } else {
      // Add items to the grid based on the saved state
      gridState.items.forEach((itemData) => {
        const newItem = document.createElement("div");
        newItem.className = "item";
        newItem.innerHTML = itemData.innerHTML.trim();
        var scriptTag = document.createElement("script");
        scriptTag.type = "text/javascript";
        scriptTag.defer = true;
        scriptTag.innerHTML = itemData.widgetScript;
        newItem.appendChild(scriptTag);
        const item = grid.add(newItem);
      });
    }

    // Layout the grid
    grid.layout(gridState.layout);
  } catch (error) {
    console.error("Error loading grid state:", error);
  }
}

// Editable canvas is used to make the canvas editable allowing the user to add widgets and be able to move the
// widgets around the canvas
function editGrid(grid, layout) {
  removeNoWidgetPlaceholder();
  try {
    if (grid) {
      grid.destroy(true);
    }
    grid = new Muuri(gridClass, {
      dragEnabled: true,
      dragHandle: ".item-content",
    });
    if (layout) {
      const gridState = JSON.parse(layout);

      if (gridState) {
        gridState.items.forEach((itemData) => {
          const newItem = document.createElement("div");
          newItem.className = "item";
          newItem.classList.add("item-editable");
          newItem.innerHTML = itemData.innerHTML.trim();
          var scriptTag = document.createElement("script");
          scriptTag.type = "text/javascript";
          scriptTag.defer = true;
          scriptTag.innerHTML = itemData.widgetScript;
          newItem.appendChild(scriptTag);
          grid.add(newItem);
          resizeObserver.observe(newItem);
        });
        grid.layout(gridState.layout);
      }
    }
  } catch (error) {
    console.error("Error loading grid state:", error);
  }

  document.getElementById("editableCanvasButton").classList.add("display-none");
  document.getElementById("CanvasButtons").classList.remove("display-none");
  document.querySelectorAll("#removeItem").forEach((item) => {
    item.classList.remove("no-opacity");
    item.disabled = false;
  });

  return grid;
}

if(window.ResizeObserver) {
  const h1Elem = document.querySelector('h1');
  const pElem = document.querySelector('p');
  const divElem = document.querySelector('body > div');

  divElem.style.width = '500px';
}
const previousSizes = new Map();

const resizeObserver = new ResizeObserver((entries) => {
  for (let entry of entries) {
    const { target } = entry;
    const previousSize = previousSizes.get(target) || {
      width: target.clientWidth,
      height: target.clientHeight,
    };
    var item = grid.getItems(target)[0];
    var el = item.getElement();
    grid = item.getGrid();
    const contentEl = el.querySelector(".item-content");
    // Calculate the change in width and height
    var widthChange = target.clientWidth - previousSize.width;
    var heightChange = target.clientHeight - previousSize.height;
    var itemContentWidth =
      parseInt(window.getComputedStyle(contentEl).getPropertyValue("width")) + widthChange;
    var itemContentHeight =
      parseInt(window.getComputedStyle(contentEl).getPropertyValue("height")) + heightChange;

    // Update the previous size for the next callback
    previousSizes.set(target, {
      width: target.clientWidth,
      height: target.clientHeight,
    });

    el.style.width = target.clientWidth + "px";
    el.style.height = target.clientHeight + "px";
    el.querySelector(".item-content").style.width = itemContentWidth + "px";
    el.querySelector(".item-content").style.height = itemContentHeight + "px";

    var chart = echarts.getInstanceByDom(contentEl);
    if (chart) {
      chart.resize({
        width: itemContentWidth,
        height: itemContentHeight,
        });
    }
    grid.refreshItems();
    grid.layout(true);
  }
});

// Save the grid layout
function saveGrid(grid, dashboardID) {
  grid._settings.dragEnabled = false;
  document.querySelectorAll("#removeItem").forEach((item) => {
    item.classList.add("no-opacity");
    item.disabled = true;
  });
  saveLayout(grid, dashboardID);
}

// Cancel the grid layout
function cancelEditGrid(grid) {
  grid._settings.dragEnabled = false;
  window.location.reload();
}

// No widget placeholder
function showNoWidgetPlaceholder() {
  const noWidgetPlaceholder = document.querySelector(".no-widget-placeholder");
  const newPlaceholder = document.createElement("div");
  newPlaceholder.innerHTML = `
  <div class="row d-flex justify-content-center">
    <div class="col-lg-4 no-widget-box text-center fs-2 px-0">
      <button
        type="button"
        class="no-widget-button w-100 p-3"
        data-bs-toggle="offcanvas"
        data-bs-target="#widgetsCanvas"
        aria-controls="widgetsCanvas"
        onclick="editableCanvas()"
      >
        <i class="fas fa-plus"></i>
        <span>Add Widgets</span>
      </button>
    </div>
  </div>
  `;

  noWidgetPlaceholder.appendChild(newPlaceholder);
}

function removeNoWidgetPlaceholder() {
  const noWidgetPlaceholder = document.querySelector(".no-widget-placeholder");
  noWidgetPlaceholder.remove();
}
