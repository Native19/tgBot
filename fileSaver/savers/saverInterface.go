package savers

type Saver interface {
	GetToDoList(chatID int64) ([]byte, error)
	RemoveToDoList(chatID int64) error
	SaveInToToDoList(chatID int64, userName string, message string) error
}
