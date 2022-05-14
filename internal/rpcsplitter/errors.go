package rpcsplitter

import "strings"

type errorList []error

func (e errorList) Error() string {
	switch len(e) {
	case 0:
		return "unknown error"
	case 1:
		return e[0].Error()
	default:
		s := strings.Builder{}
		s.WriteString("the following errors occurred: ")
		s.WriteString("[")
		for n, err := range e {
			s.WriteString(err.Error())
			if n < len(e)-1 {
				s.WriteString(", ")
			}
		}
		s.WriteString("]")
		return s.String()
	}
}

// addError adds an error to an error slice. If errs is not an error slice it
// will be converted into one. If there is already an error with the same
// message in the slice, it will not be added.
func addError(err error, errs ...error) error {
	if err == nil {
		return nil
	}
	if _, ok := err.(errorList); !ok {
		err = errorList{err}
	}
	r := err.(errorList)
	for _, e := range errs {
		if e == nil {
			continue
		}
		r = append(r, e)
	}
	return r
}
