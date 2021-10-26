package projects

import (
	log "github.com/Sirupsen/logrus"
	"github.com/ansible-semaphore/semaphore/api/helpers"
	"github.com/ansible-semaphore/semaphore/db"
	"net/http"

	"github.com/gorilla/context"
)

// ViewMiddleware ensures a key exists and loads it to the context
func ViewMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		project := context.Get(r, "project").(db.Project)
		viewID, err := helpers.GetIntParam("view_id", w, r)
		if err != nil {
			return
		}

		view, err := helpers.Store(r).GetView(project.ID, viewID)

		if err != nil {
			helpers.WriteError(w, err)
			return
		}

		context.Set(r, "view", view)
		next.ServeHTTP(w, r)
	})
}

// GetViews retrieves sorted keys from the database
func GetViews(w http.ResponseWriter, r *http.Request) {
	if view := context.Get(r, "view"); view != nil {
		k := view.(db.View)
		helpers.WriteJSON(w, http.StatusOK, k)
		return
	}

	project := context.Get(r, "project").(db.Project)
	var views []db.View

	views, err := helpers.Store(r).GetViews(project.ID)

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, views)
}

// AddView adds a new key to the database
func AddView(w http.ResponseWriter, r *http.Request) {
	project := context.Get(r, "project").(db.Project)
	var view db.View

	if !helpers.Bind(w, r, &view) {
		return
	}

	if view.ProjectID != project.ID {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Project ID in body and URL must be the same",
		})
		return
	}

	newView, err := helpers.Store(r).CreateView(view)

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	user := context.Get(r, "user").(*db.User)

	objType := db.EventKey

	desc := "View " + view.Title + " created"
	_, err = helpers.Store(r).CreateEvent(db.Event{
		UserID:      &user.ID,
		ProjectID:   &newView.ProjectID,
		ObjectType:  &objType,
		ObjectID:    &newView.ID,
		Description: &desc,
	})

	if err != nil {
		log.Error(err)
	}

	w.WriteHeader(http.StatusNoContent)
}

// UpdateView updates key in database
// nolint: gocyclo
func UpdateView(w http.ResponseWriter, r *http.Request) {
	var view db.View
	oldView := context.Get(r, "view").(db.View)

	if !helpers.Bind(w, r, &view) {
		return
	}

	if err := helpers.Store(r).UpdateView(view); err != nil {
		helpers.WriteError(w, err)
		return
	}

	user := context.Get(r, "user").(*db.User)

	desc := "View " + view.Title + " updated"
	objType := db.EventView

	_, err := helpers.Store(r).CreateEvent(db.Event{
		UserID:      &user.ID,
		ProjectID:   &oldView.ProjectID,
		Description: &desc,
		ObjectID:    &oldView.ID,
		ObjectType:  &objType,
	})

	if err != nil {
		log.Error(err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// RemoveView deletes a view from the database
func RemoveView(w http.ResponseWriter, r *http.Request) {
	view := context.Get(r, "view").(db.View)

	var err error

	err = helpers.Store(r).DeleteView(view.ProjectID, view.ID)

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	user := context.Get(r, "user").(*db.User)

	desc := "View " + view.Title + " deleted"

	_, err = helpers.Store(r).CreateEvent(db.Event{
		UserID:      &user.ID,
		ProjectID:   &view.ProjectID,
		Description: &desc,
	})

	if err != nil {
		log.Error(err)
	}

	w.WriteHeader(http.StatusNoContent)
}